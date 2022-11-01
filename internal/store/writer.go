package store

import (
	"context"

	"github.com/nicolastakashi/jaeger-redisearch/internal/repository"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
)

type SpanWriter struct {
	logger              hclog.Logger
	spanRepository      *repository.SpanRepository
	operationRepository *repository.OperationRepository
}

func NewSpanWriter(logger hclog.Logger, spanRepository *repository.SpanRepository, serviceRepository *repository.OperationRepository) *SpanWriter {
	return &SpanWriter{
		logger:              logger,
		spanRepository:      spanRepository,
		operationRepository: serviceRepository,
	}
}

func (s *SpanWriter) WriteSpan(ctx context.Context, span *jModel.Span) error {
	err := s.operationRepository.Write(ctx, span)

	if err != nil {
		s.logger.Error("error to write service", err)
		return err
	}

	err = s.spanRepository.Write(ctx, span)
	if err != nil {
		s.logger.Error("error to write span", err)
		return err
	}
	return nil
}
