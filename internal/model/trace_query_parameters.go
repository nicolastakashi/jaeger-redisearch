package model

import "time"

type TraceQueryParameters struct {
	ServiceName   string
	OperationName string
	Tags          map[string]string
	StartTimeMin  time.Time
	StartTimeMax  time.Time
	DurationMin   time.Duration
	DurationMax   time.Duration
	NumTraces     int64
}
