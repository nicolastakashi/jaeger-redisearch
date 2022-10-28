package repository

import (
	"context"
	"fmt"
	"jaegerredissearch/internal/model"
	"jaegerredissearch/internal/redis"
	"strings"

	"github.com/hashicorp/go-hclog"
	jModel "github.com/jaegertracing/jaeger/model"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/om"
)

type SpanRepository struct {
	logger     hclog.Logger
	repository om.Repository[model.Span]
}

func NewSpanRepository(logger hclog.Logger, redisClient rueidis.Client) (*SpanRepository, error) {
	repository := om.NewJSONRepository("spans", model.Span{}, redisClient)
	if _, ok := repository.(*om.JSONRepository[model.Span]); ok {
		createSpanIndex(repository)
	}
	return &SpanRepository{
		logger:     logger,
		repository: repository,
	}, nil
}

func createSpanIndex(repository om.Repository[model.Span]) {
	repository.CreateIndex(context.TODO(), func(schema om.FtCreateSchema) om.Completed {
		text := schema.FieldName("$.traceID").As("traceID").Text()
		text = text.FieldName("$.spanID").As("spanID").Text()
		text = text.FieldName("$.operationName").As("operationName").Text()

		text = text.FieldName("$.process.serviceName").As("processServiceName").Text()

		tag := text.FieldName("$.process.tags[0:].key").As("processTagKey").Tag()
		tag = tag.FieldName("$.process.tags[0:].type").As("processTagType").Tag()
		tag = tag.FieldName("$.process.tags[0:].value").As("processTagValue").Tag()

		tag = tag.FieldName("$.tags[0:].key").As("tagKey").Tag()
		tag = tag.FieldName("$.tags[0:].type").As("tagType").Tag()
		tag = tag.FieldName("$.tags[0:].value").As("tagValue").Tag()

		tag = tag.FieldName("$.references[0:].refType").As("refType").Tag()
		tag = tag.FieldName("$.references[0:].traceID").As("refTraceID").Tag()
		tag = tag.FieldName("$.references[0:].spanID").As("refSpanID").Tag()

		numeric := tag.FieldName("$.startTime").As("startTime").Numeric()
		numeric = numeric.FieldName("$.flags").As("flags").Numeric()
		numeric = numeric.FieldName("$.duration").As("duration").Numeric()

		return numeric.Build()
	})
}

func (s *SpanRepository) WriteSpan(context context.Context, jSpan *jModel.Span) error {
	span := s.repository.NewEntity()

	span.TraceID = jSpan.TraceID.String()
	span.SpanID = jSpan.SpanID.String()
	span.OperationName = redis.Tokenization(jSpan.OperationName)
	span.StartTime = model.TimeAsEpochMicroseconds(jSpan.StartTime)
	span.Duration = model.DurationAsMicroseconds(jSpan.Duration)
	span.References = model.ConvertReferencesFromJaeger(jSpan)
	span.ProcessID = jSpan.ProcessID
	span.Process = model.ConvertProcessFromJager(jSpan.Process)
	span.Tags = model.ConvertKeyValuesFromJaeger(jSpan.Tags)
	span.Warnings = jSpan.Warnings

	err := s.repository.Save(context, span)

	if err != nil {
		return err
	}

	return nil
}

func (s *SpanRepository) GetTracesId(context context.Context, service string) ([]string, error) {
	cursor, err := s.repository.Aggregate(context, func(search om.FtAggregateIndex) om.Completed {
		query := fmt.Sprintf("@processServiceName:%s", redis.Tokenization(service))
		return search.Query(query).LoadAll().Groupby(1).Property("@traceID").Reduce("COUNT").Nargs(0).Build()
	})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	total := cursor.Total()
	services := make([]string, total)

	c, err := cursor.Read(context)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	for i, s := range c {
		services[i] = s["traceID"]
	}

	return services, nil
}

func (s *SpanRepository) GetTracesById(context context.Context, ids []string) (map[string]*jModel.Trace, error) {
	_, spans, err := s.repository.Search(context, func(search om.FtSearchIndex) om.Completed {
		query := fmt.Sprintf("@traceID:(%s)", strings.Join(ids, "|"))
		return search.Query(query).Build()
	})

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	tracesMap := make(map[string]*jModel.Trace, len(ids))
	for _, span := range spans {
		if _, ok := tracesMap[span.TraceID]; !ok {
			tracesMap[span.TraceID] = &jModel.Trace{}
		}

		tId, _ := jModel.TraceIDFromString(span.TraceID)
		sId, _ := jModel.SpanIDFromString(span.SpanID)
		refs, _ := model.ConvertReferencesToJaeger(span.References)
		tags, _ := model.ConvertKeyValuesToJaeger(span.Tags)
		pTags, _ := model.ConvertKeyValuesToJaeger(span.Process.Tags)

		s := jModel.Span{
			TraceID:    tId,
			SpanID:     sId,
			References: refs,
			Tags:       tags,
			StartTime:  jModel.EpochMicrosecondsAsTime(span.StartTime),
			Duration:   jModel.MicrosecondsAsDuration(span.Duration),
			Process: &jModel.Process{
				ServiceName: redis.UnTokenization(span.Process.ServiceName),
				Tags:        pTags,
			},
		}

		tracesMap[span.TraceID].Spans = append(tracesMap[span.TraceID].Spans, &s)
	}
	return tracesMap, nil
}

func (s *SpanRepository) GetSpans(context context.Context, service string) ([]*model.Span, error) {
	_, records, err := s.repository.Search(context, func(search om.FtSearchIndex) om.Completed {
		query := fmt.Sprintf("@processServiceName:%s", redis.Tokenization(service))
		return search.Query(query).Build()
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}
