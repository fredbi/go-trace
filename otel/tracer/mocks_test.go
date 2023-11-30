package tracer

import (
	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/otel/exporters/mock"
	"go.opentelemetry.io/otel/trace"
)

type mockRuntime struct {
	tracer *mock.Tracer
	logger log.Factory
}

func (m mockRuntime) Logger() log.Factory {
	return m.logger
}
func (m mockRuntime) Tracer() trace.Tracer {
	return m.tracer
}
