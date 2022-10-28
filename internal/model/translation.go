package model

import (
	"fmt"
	"jaegerredissearch/internal/redis"
	"strconv"
	"strings"
	"time"

	jModel "github.com/jaegertracing/jaeger/model"
)

func ConvertProcessFromJager(process *jModel.Process) Process {
	return Process{
		ServiceName: redis.Tokenization(process.ServiceName),
		Tags:        ConvertKeyValuesFromJaeger(process.Tags),
	}
}

func ConvertReferencesFromJaeger(span *jModel.Span) []Reference {
	out := make([]Reference, 0, len(span.References))
	for _, ref := range span.References {
		out = append(out, Reference{
			RefType: ConvertRefTypeFromJaeger(ref.RefType),
			TraceID: TraceID(ref.TraceID.String()),
			SpanID:  SpanID(ref.SpanID.String()),
		})
	}
	return out
}

func ConvertRefTypeFromJaeger(refType jModel.SpanRefType) ReferenceType {
	if refType == jModel.FollowsFrom {
		return FollowsFrom
	}
	return ChildOf
}

func ConvertKeyValuesFromJaeger(keyValues jModel.KeyValues) []KeyValue {
	kvs := make([]KeyValue, 0)
	for _, kv := range keyValues {
		if kv.GetVType() != jModel.BinaryType {
			kvs = append(kvs, convertKeyValueFromJaeger(kv))
		}
	}
	return kvs
}

func convertKeyValueFromJaeger(kv jModel.KeyValue) KeyValue {
	return KeyValue{
		Key:   redis.Tokenization(kv.Key),
		Type:  ValueType(strings.ToLower(kv.VType.String())),
		Value: redis.Tokenization(kv.AsString()),
	}
}

func TimeAsEpochMicroseconds(t time.Time) uint64 {
	return uint64(t.UnixNano() / 1000)
}

func DurationAsMicroseconds(d time.Duration) uint64 {
	return uint64(d.Nanoseconds() / 1000)
}

func ConvertReferencesToJaeger(refs []Reference) ([]jModel.SpanRef, error) {
	retMe := make([]jModel.SpanRef, len(refs))
	for i, r := range refs {
		// There are some inconsistencies with ReferenceTypes, hence the hacky fix.
		var refType jModel.SpanRefType
		switch r.RefType {
		case ChildOf:
			refType = jModel.ChildOf
		case FollowsFrom:
			refType = jModel.FollowsFrom
		default:
			return nil, fmt.Errorf("not a valid SpanRefType string %s", string(r.RefType))
		}

		traceID, err := jModel.TraceIDFromString(string(r.TraceID))
		if err != nil {
			return nil, err
		}

		spanID, err := jModel.SpanIDFromString(string(r.SpanID))
		if err != nil {
			return nil, err
		}

		retMe[i] = jModel.SpanRef{
			RefType: refType,
			TraceID: traceID,
			SpanID:  spanID,
		}
	}
	return retMe, nil
}

func ConvertKeyValuesToJaeger(tags []KeyValue) ([]jModel.KeyValue, error) {
	retMe := make([]jModel.KeyValue, len(tags))
	for i := range tags {
		kv, err := convertKeyValueToJaeger(&tags[i])
		if err != nil {
			return nil, err
		}
		retMe[i] = kv
	}
	return retMe, nil
}

func convertKeyValueToJaeger(tag *KeyValue) (jModel.KeyValue, error) {
	if tag.Value == nil {
		return jModel.KeyValue{}, fmt.Errorf("invalid nil Value in %v", tag)
	}
	tagValue, ok := tag.Value.(string)
	if !ok {
		return jModel.KeyValue{}, fmt.Errorf("non-string Value of type %t in %v", tag.Value, tag)
	}
	switch tag.Type {
	case StringType:
		return jModel.String(redis.UnTokenization(tag.Key), redis.UnTokenization(tagValue)), nil
	case BoolType:
		value, err := strconv.ParseBool(tagValue)
		if err != nil {
			return jModel.KeyValue{}, err
		}
		return jModel.Bool(redis.UnTokenization(tag.Key), value), nil
	case Int64Type:
		value, err := strconv.ParseInt(tagValue, 10, 64)
		if err != nil {
			return jModel.KeyValue{}, err
		}
		return jModel.Int64(redis.UnTokenization(tag.Key), value), nil
	case Float64Type:
		value, err := strconv.ParseFloat(tagValue, 64)
		if err != nil {
			return jModel.KeyValue{}, err
		}
		return jModel.Float64(redis.UnTokenization(tag.Key), value), nil
	}
	return jModel.KeyValue{}, fmt.Errorf("not a valid ValueType string %s", string(tag.Type))
}
