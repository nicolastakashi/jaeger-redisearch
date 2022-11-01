package model

// ReferenceType is the reference type of one span to another
type ReferenceType string

// TraceID is the shared trace ID of all spans in the trace.
type TraceID string

// SpanID is the id of a span
type SpanID string

// ValueType is the type of a value stored in KeyValue struct.
type ValueType string

const (
	// ChildOf means a span is the child of another span
	ChildOf ReferenceType = "CHILD_OF"
	// FollowsFrom means a span follows from another span
	FollowsFrom ReferenceType = "FOLLOWS_FROM"

	// StringType indicates a string value stored in KeyValue
	StringType ValueType = "string"
	// BoolType indicates a Boolean value stored in KeyValue
	BoolType ValueType = "bool"
	// Int64Type indicates a 64bit signed integer value stored in KeyValue
	Int64Type ValueType = "int64"
	// Float64Type indicates a 64bit float value stored in KeyValue
	Float64Type ValueType = "float64"
	// BinaryType indicates a binary value stored in KeyValue
	BinaryType ValueType = "binary"
)

// Span is MongoDB representation of the domain span.
type Span struct {
	Key           string      `json:"key" redis:",key"` // the redis:",key" is required to indicate which field is the ULID key
	Ver           int64       `json:"ver" redis:",ver"` // the redis:",ver" is required to do optimistic locking to prevent lost update
	TraceID       string      `json:"traceID"`
	SpanID        string      `json:"spanID"`
	OperationName string      `json:"operationName"`
	StartTime     uint64      `json:"startTime"` // microseconds since Unix epoch
	Duration      uint64      `json:"duration"`  // microseconds
	References    []Reference `json:"references"`
	ProcessID     string      `json:"processID"`
	Process       Process     `json:"process,omitempty"`
	Tags          []KeyValue  `json:"tags"`
	MultipleTags  []KeyValue  `json:"mTags"`
	Logs          []Log       `json:"logs"`
	Warnings      []string    `json:"warnings"`
}

type Reference struct {
	RefType ReferenceType `json:"refType"`
	TraceID TraceID       `json:"traceID"`
	SpanID  SpanID        `json:"spanID"`
}

type Process struct {
	ServiceName string     `json:"serviceName"`
	Tags        []KeyValue `json:"tags"`
}

type Log struct {
	Timestamp uint64     `bson:"timestamp"`
	Fields    []KeyValue `bson:"fields"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Type  ValueType   `json:"type,omitempty"`
	Value interface{} `json:"value"`
}
