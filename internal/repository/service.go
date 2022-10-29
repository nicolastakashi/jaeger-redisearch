package repository

import (
	"context"
	"fmt"
	"hash/fnv"
	"jaegerredissearch/internal/metrics"
	"jaegerredissearch/internal/model"
	"jaegerredissearch/internal/redis"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/om"
)

const serviceIndexName = "services"

type ServiceRepository struct {
	logger     hclog.Logger
	repository om.Repository[model.Service]
	mu         sync.Mutex
}

func NewServiceRepository(logger hclog.Logger, redisClient rueidis.Client) (*ServiceRepository, error) {
	repository := om.NewJSONRepository(serviceIndexName, model.Service{}, redisClient)
	if _, ok := repository.(*om.JSONRepository[model.Service]); ok {
		createServiceIndex(repository)
	}
	return &ServiceRepository{
		logger:     logger,
		repository: repository,
		mu:         sync.Mutex{},
	}, nil
}

func createServiceIndex(repository om.Repository[model.Service]) {
	repository.CreateIndex(context.TODO(), func(schema om.FtCreateSchema) om.Completed {
		text := schema.FieldName("$.service").As("service").Text()
		text = text.FieldName("$.operation").As("operation").Text()
		text = text.FieldName("$.hash").As("hash").Text()
		return text.Build()
	})
}

func (s *ServiceRepository) WriteService(context context.Context, jaegerSpan *jModel.Span) error {
	s.mu.Lock()
	writeStart := time.Now()
	defer s.mu.Unlock()

	hash := hashCode(jaegerSpan)

	n, _, err := s.repository.Search(context, func(search om.FtSearchIndex) om.Completed {
		return search.Query(hash).Build()
	})

	if err != nil {
		return err
	}

	if n > 0 {
		return nil
	}

	newSvc := s.repository.NewEntity()
	newSvc.ServiceName = redis.Tokenization(jaegerSpan.Process.ServiceName)
	newSvc.OperationName = redis.Tokenization(jaegerSpan.OperationName)
	newSvc.Hash = hash

	err = s.repository.Save(context, newSvc)

	if err != nil {
		metrics.WritesLantency.WithLabelValues(serviceIndexName, "Error").Observe(time.Since(writeStart).Seconds())
		return err
	}

	metrics.WritesLantency.WithLabelValues(serviceIndexName, "Ok").Observe(time.Since(writeStart).Seconds())
	metrics.WritesTotal.WithLabelValues(serviceIndexName).Inc()

	return nil
}

func (s *ServiceRepository) GetServices(context context.Context) ([]string, error) {
	cursor, err := s.repository.Aggregate(context, func(search om.FtAggregateIndex) om.Completed {
		return search.Query("*").LoadAll().Groupby(1).Property("@service").Reduce("COUNT").Nargs(0).Build()
	})

	if err != nil {
		return nil, err
	}

	total := cursor.Total()
	services := make([]string, total)

	c, err := cursor.Read(context)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	for i, s := range c {
		services[i] = redis.UnTokenization(s["service"])
	}

	return services, nil
}

func (s *ServiceRepository) GetServiceOperation(context context.Context, service string) ([]string, error) {
	n, records, err := s.repository.Search(context, func(search om.FtSearchIndex) om.Completed {
		query := fmt.Sprintf("@service:%s", redis.Tokenization(service))
		return search.Query(query).Build()
	})

	if err != nil {
		return nil, err
	}

	operations := make([]string, n)

	for i, r := range records {
		operations[i] = redis.UnTokenization(r.OperationName)
	}

	return operations, nil
}

func hashCode(jaegerSpan *jModel.Span) string {
	h := fnv.New64a()
	h.Write([]byte(jaegerSpan.Process.ServiceName))
	h.Write([]byte(jaegerSpan.OperationName))
	return fmt.Sprintf("%x", h.Sum64())
}
