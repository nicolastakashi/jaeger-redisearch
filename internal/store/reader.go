package store

import (
	"context"
	"fmt"
	"jaegerredissearch/internal/model"
	"jaegerredissearch/internal/repository"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

type SpanReader struct {
	logger            hclog.Logger
	spanRepository    *repository.SpanRepository
	serviceRepository *repository.ServiceRepository
}

func NewSpanReader(logger hclog.Logger, spanRepository *repository.SpanRepository, serviceRepository *repository.ServiceRepository) *SpanReader {
	return &SpanReader{
		logger:            logger,
		spanRepository:    spanRepository,
		serviceRepository: serviceRepository,
	}
}

func (s *SpanReader) GetServices(ctx context.Context) ([]string, error) {
	services, err := s.serviceRepository.GetServices(ctx)

	if err != nil {
		return nil, fmt.Errorf("error to get services: %s", err)
	}

	return services, nil
}

func (s *SpanReader) GetTrace(ctx context.Context, traceID jModel.TraceID) (*jModel.Trace, error) {
	return nil, fmt.Errorf("xablau")
}

func (s *SpanReader) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	operations, err := s.serviceRepository.GetServiceOperation(ctx, query.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("error to get services: %s", err)
	}
	array := make([]spanstore.Operation, len(operations))
	for i, s := range operations {
		array[i] = spanstore.Operation{
			Name: s,
		}
	}
	return array, nil
}

func (s *SpanReader) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*jModel.Trace, error) {
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
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("error to get traces id: %s", err)
	}

	tracesMap, err := s.spanRepository.GetTracesById(ctx, traceIds)

	if err != nil {
		return nil, fmt.Errorf("error to get traces by id: %s", err)
	}

	var traces []*jModel.Trace
	for _, trace := range tracesMap {
		traces = append(traces, trace)
	}

	return traces, nil
}

func (s *SpanReader) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]jModel.TraceID, error) {
	return nil, nil
}
