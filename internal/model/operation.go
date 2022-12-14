package model

type Operation struct {
	Key           string `json:"key" redis:",key"` // the redis:",key" is required to indicate which field is the ULID key
	Ver           int64  `json:"ver" redis:",ver"` // the redis:",ver" is required to do optimistic locking to prevent lost update
	ServiceName   string `json:"service"`
	OperationName string `json:"operation"`
	SpanKind      string `json:"span_kind"`
	Hash          string `json:"hash"`
}
