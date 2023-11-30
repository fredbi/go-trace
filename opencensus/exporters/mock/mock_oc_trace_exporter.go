package mock

import (
	"encoding/json"
	"sync"
	"testing"

	"go.opencensus.io/trace"
)

var _ trace.Exporter = &Exporter{}

// Exporter implements a mock for the Opencensus trace exporter
type Exporter struct {
	t        testing.TB
	Exported []*trace.SpanData
	count    int
	mx       sync.Mutex
}

// New builds a new Opencensus mock trace exporter
func New(t testing.TB) *Exporter {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	exporter := &Exporter{t: t}
	trace.RegisterExporter(exporter)

	return exporter
}

func (r *Exporter) ExportSpan(s *trace.SpanData) {
	b, _ := json.Marshal(s)
	r.t.Logf("%s", string(b))
	r.mx.Lock()
	r.count++
	r.Exported = append(r.Exported, s)
	r.mx.Unlock()
}

// Count the number of times the exporter was invoked.
func (r *Exporter) Count() int {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.count
}
