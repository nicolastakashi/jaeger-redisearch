package repository

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/nicolastakashi/jaeger-redisearch/internal/metrics"
	"github.com/nicolastakashi/jaeger-redisearch/internal/model"
	"github.com/nicolastakashi/jaeger-redisearch/internal/redis"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/om"
)

const operationIndexName = "operation"

type OperationRepository struct {
	logger     hclog.Logger
	repository om.Repository[model.Operation]
	mu         sync.Mutex
	client     rueidis.Client
	config     model.Configuration
}

func NewOperationRepository(logger hclog.Logger, redisClient rueidis.Client, config model.Configuration) (*OperationRepository, error) {
	repository := om.NewJSONRepository(operationIndexName, model.Operation{}, redisClient)
	if _, ok := repository.(*om.JSONRepository[model.Operation]); ok {
		createOperationIndex(repository)
	}
	return &OperationRepository{
		logger:     logger,
		repository: repository,
		mu:         sync.Mutex{},
		client:     redisClient,
		config:     config,
	}, nil
}

func createOperationIndex(repository om.Repository[model.Operation]) {
	repository.CreateIndex(context.TODO(), func(schema om.FtCreateSchema) om.Completed {
		text := schema.FieldName("$.service").As("service").Text()
		text = text.FieldName("$.operation").As("operation").Text()
		text = text.FieldName("$.span_kind").As("span_kind").Text()
		text = text.FieldName("$.hash").As("hash").Text()
		return text.Build()
	})
}

func (s *OperationRepository) Write(context context.Context, jaegerSpan *jModel.Span) error {
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

	spanKind := ""
	for _, tag := range jaegerSpan.Tags {
		if tag.Key == "span.kind" {
			spanKind = tag.AsString()
		}
	}

	newSvc := s.repository.NewEntity()
	newSvc.ServiceName = redis.Tokenization(jaegerSpan.Process.ServiceName)
	newSvc.OperationName = redis.Tokenization(jaegerSpan.OperationName)
	newSvc.SpanKind = spanKind
	newSvc.Hash = hash

	err = s.repository.Save(context, newSvc)

	if err != nil {
		metrics.WritesLantency.WithLabelValues(operationIndexName, "Error").Observe(time.Since(writeStart).Seconds())
		return err
	}

	setTTL(context, s.client, fmt.Sprintf("%v:%v", operationIndexName, newSvc.Key), s.config.RedisTTL)

	metrics.WritesLantency.WithLabelValues(operationIndexName, "Ok").Observe(time.Since(writeStart).Seconds())
	metrics.WritesTotal.WithLabelValues(operationIndexName).Inc()

	return nil
}

func (s *OperationRepository) GetServices(context context.Context) ([]string, error) {
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

func (s *OperationRepository) GetOperationsByService(context context.Context, service string) ([]*model.Operation, error) {
	_, records, err := s.repository.Search(context, func(search om.FtSearchIndex) om.Completed {
		query := fmt.Sprintf("@service:%s", redis.Tokenization(service))
		return search.Query(query).Build()
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}

func hashCode(jaegerSpan *jModel.Span) string {
	h := fnv.New64a()
	h.Write([]byte(jaegerSpan.Process.ServiceName))
	h.Write([]byte(jaegerSpan.OperationName))
	return fmt.Sprintf("%x", h.Sum64())
}

func setTTL(context context.Context, client rueidis.Client, key string, ttl time.Duration) {
	client.Do(context, client.B().Expire().Key(key).Seconds(int64(ttl.Seconds())).Build())
}
