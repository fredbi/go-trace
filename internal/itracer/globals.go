package itracer

import (
	"sync"
)

var (
	prefixLock sync.Mutex
	prefix     = "function"
)

// RegisterPrefix sets a package level tracer prefix at initialization time.
//
// This is used as the key in structured logs to hold the signature of the trace.
//
// The default value is "function", so a log entry looks like:
//
//	2023-11-01T17:19:58.615+0100	INFO	tracer/example_test.go:33	test	{
//		"function": "tracer_test.ExampleStartSpan",
//		"field": "fred"
//		}
func RegisterPrefix(custom string) {
	prefixLock.Lock()
	prefix = custom
	prefixLock.Unlock()
}
