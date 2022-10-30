package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nicolastakashi/jaeger-redisearch/internal/metrics"
	"github.com/nicolastakashi/jaeger-redisearch/internal/model"
	"github.com/nicolastakashi/jaeger-redisearch/internal/redis"
	"github.com/nicolastakashi/jaeger-redisearch/internal/repository"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

type SpanReader struct {
	logger            hclog.Logger
	spanRepository    *repository.SpanRepository
	serviceRepository *repository.OperationRepository
}

func NewSpanReader(logger hclog.Logger, spanRepository *repository.SpanRepository, serviceRepository *repository.OperationRepository) *SpanReader {
	return &SpanReader{
		logger:            logger,
		spanRepository:    spanRepository,
		serviceRepository: serviceRepository,
	}
}

func (s *SpanReader) GetServices(ctx context.Context) ([]string, error) {
	defer metrics.ReadsTotal.WithLabelValues("services", "get_services")
	start := time.Now()

	services, err := s.serviceRepository.GetServices(ctx)

	if err != nil {
		metrics.ReadLatency.WithLabelValues("services", "Error", "get_services").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get services: %s", err)
	}

	metrics.ReadLatency.WithLabelValues("services", "Ok", "get_services").Observe(time.Since(start).Seconds())

	return services, nil
}

func (s *SpanReader) GetTrace(ctx context.Context, traceID jModel.TraceID) (*jModel.Trace, error) {
	defer metrics.ReadsTotal.WithLabelValues("spans", "get_trace")
	start := time.Now()

	tracesMap, err := s.spanRepository.GetTracesById(ctx, []string{traceID.String()})

	if err != nil {
		metrics.ReadLatency.WithLabelValues("spans", "Error", "get_trace").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get traces by id: %s", err)
	}

	for _, trace := range tracesMap {
		metrics.ReadLatency.WithLabelValues("spans", "Ok", "get_trace").Observe(time.Since(start).Seconds())
		return trace, nil
	}

	metrics.ReadLatency.WithLabelValues("spans", "Error", "get_trace").Observe(time.Since(start).Seconds())
	return nil, errors.New("trace not found")
}

func (s *SpanReader) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	defer metrics.ReadsTotal.WithLabelValues("services", "get_operations")
	start := time.Now()

	operations, err := s.serviceRepository.GetOperationsByService(ctx, query.ServiceName)
	if err != nil {
		metrics.ReadLatency.WithLabelValues("services", "Error", "get_operations").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get services: %s", err)
	}

	array := make([]spanstore.Operation, len(operations))
	for i, operation := range operations {
		array[i] = spanstore.Operation{
			Name:     redis.UnTokenization(operation.OperationName),
			SpanKind: operation.SpanKind,
		}
	}

	metrics.ReadLatency.WithLabelValues("services", "Ok", "get_operations").Observe(time.Since(start).Seconds())
	return array, nil
}

func (s *SpanReader) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*jModel.Trace, error) {
	defer metrics.ReadsTotal.WithLabelValues("spans", "find_traces")
	start := time.Now()

	traceIds, err := s.spanRepository.GetTracesId(ctx, model.TraceQueryParameters{
		ServiceName:   query.ServiceName,
		OperationName: query.OperationName,
		Tags:          query.Tags,
		StartTimeMin:  query.StartTimeMin,
		StartTimeMax:  query.StartTimeMax,
		DurationMin:   query.DurationMin,
		DurationMax:   query.DurationMax,
		NumTraces:     int64(query.NumTraces),
	})

	if len(traceIds) == 0 {
		metrics.ReadLatency.WithLabelValues("spans", "Error", "find_traces").Observe(time.Since(start).Seconds())
		return nil, nil
	}

	if err != nil {
		metrics.ReadLatency.WithLabelValues("spans", "Ok", "find_traces").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get traces id: %s", err)
	}

	tracesMap, err := s.spanRepository.GetTracesById(ctx, traceIds)

	if err != nil {
		metrics.ReadLatency.WithLabelValues("spans", "Error", "find_traces").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get traces by id: %s", err)
	}

	var traces []*jModel.Trace
	for _, trace := range tracesMap {
		traces = append(traces, trace)
	}

	metrics.ReadLatency.WithLabelValues("spans", "Ok", "find_traces").Observe(time.Since(start).Seconds())
	return traces, nil
}

func (s *SpanReader) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]jModel.TraceID, error) {
	defer metrics.ReadsTotal.WithLabelValues("spans", "find_trace_ids")
	start := time.Now()

	traceIds, err := s.spanRepository.GetTracesId(ctx, model.TraceQueryParameters{
		ServiceName:   query.ServiceName,
		OperationName: query.OperationName,
		Tags:          query.Tags,
		StartTimeMin:  query.StartTimeMin,
		StartTimeMax:  query.StartTimeMax,
		DurationMin:   query.DurationMin,
		DurationMax:   query.DurationMax,
		NumTraces:     int64(query.NumTraces),
	})

	if len(traceIds) == 0 {
		metrics.ReadLatency.WithLabelValues("spans", "Ok", "find_trace_ids").Observe(time.Since(start).Seconds())
		return nil, nil
	}

	if err != nil {
		metrics.ReadLatency.WithLabelValues("spans", "Error", "find_trace_ids").Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("error to get traces id: %s", err)
	}

	traceIDs := make([]jModel.TraceID, len(traceIds))

	for i, id := range traceIds {
		t, err := jModel.TraceIDFromString(id)
		if err != nil {
			return nil, err
		}
		traceIDs[i] = t
	}

	metrics.ReadLatency.WithLabelValues("spans", "Ok", "find_trace_ids").Observe(time.Since(start).Seconds())
	return traceIDs, nil
}

func (s *SpanReader) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]jModel.DependencyLink, error) {
	return []jModel.DependencyLink{}, nil
}
