package store

import (
	"context"
	"jaegerredissearch/internal/repository"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
)

type SpanWriter struct {
	logger            hclog.Logger
	spanRepository    *repository.SpanRepository
	serviceRepository *repository.ServiceRepository
}

func NewSpanWriter(logger hclog.Logger, spanRepository *repository.SpanRepository, serviceRepository *repository.ServiceRepository) *SpanWriter {
	return &SpanWriter{
		logger:            logger,
		spanRepository:    spanRepository,
		serviceRepository: serviceRepository,
	}
}

func (s *SpanWriter) WriteSpan(ctx context.Context, span *jModel.Span) error {
	err := s.serviceRepository.WriteService(ctx, span)

	if err != nil {
		s.logger.Error("error to write service", err)
		return err
	}

	err = s.spanRepository.WriteSpan(ctx, span)
	if err != nil {
		s.logger.Error("error to write span", err)
		return err
	}
	return nil
}
