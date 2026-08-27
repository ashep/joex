package eventstream

import (
	"time"
)

type Streamable interface {
	StreamEventKey() string
	StreamEventValue() []byte
	StreamEventHeaders() [][2]string
	StreamEventTime() time.Time
}

type ProduceResult struct {
	Partition int
	Offset    int64
}

type ConsumedEvent[T any] struct {
	Topic     string
	Partition int
	Offset    int
	Key       []byte
	RawValue  []byte
	Value     T
}
