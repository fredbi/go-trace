// Package mock provides a mock trace exporter for testing OTEL traces.
package mock

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ sdktrace.SpanExporter = &Exporter{}
	_ trace.Tracer          = &Tracer{}
)

type (
	// Exporter implements a mock trace exporter.
	//
	// All exported spans are held in the Exported field.
	Exporter struct {
		t        testing.TB
		Exported []sdktrace.ReadOnlySpan
		count    int
		mx       sync.Mutex
	}

	// Tracer implements an OTL tracer based on a mock exporter.
	Tracer struct {
		exporter  *Exporter
		processor sdktrace.SpanProcessor

		trace.Tracer
	}
)

// New builds a mock exporter for OTEL traces
func New(t testing.TB) *Exporter {
	return &Exporter{t: t}
}

// NewTracer returns an OTEL tracer based on the mock exporter.
//
// This tracer knows how to flush exported spans.
func NewTracer(t testing.TB) *Tracer {
	const name = "mock_tracer"

	exporter := New(t)
	processor := sdktrace.NewBatchSpanProcessor(exporter)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(processor),
	)

	return &Tracer{
		exporter:  exporter,
		processor: processor,
		Tracer:    traceProvider.Tracer(name),
	}
}

func (tr *Tracer) Exporter() *Exporter {
	return tr.exporter
}

func (tr *Tracer) SpanProcessor() sdktrace.SpanProcessor {
	return tr.processor
}

func (tr *Tracer) ForceFlush(ctx context.Context) error {
	return tr.processor.ForceFlush(ctx)
}

// Count the number of times the exporter was invoked.
func (r *Exporter) Count() int {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.count
}

func (r *Exporter) ExportSpans(_ context.Context, s []sdktrace.ReadOnlySpan) error {
	b, err := json.Marshal(s)
	require.NoError(r.t, err)

	r.t.Logf("%s", string(b))
	r.mx.Lock()
	r.count++
	r.Exported = append(r.Exported, s...)
	r.mx.Unlock()

	return nil
}
func (r *Exporter) Shutdown(context.Context) error {
	return nil
}
