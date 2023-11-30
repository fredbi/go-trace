package tracer

import (
	"context"

	"github.com/fredbi/go-trace/log"
	octracer "github.com/fredbi/go-trace/opencensus/tracer"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

// Loggable is a log factory provider
type Loggable interface {
	Logger() log.Factory
}

// StartSpan returns an opencensus span and logger that prepends the caller's signature.
//
// Deprecated: use github.com/fredbi/go-tracer/opencensus/tracer.StartSpan instead.
func StartSpan(ctx context.Context, rt Loggable, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	return octracer.StartSpan(ctx, rt, fields...)
}

// StartNamedSpan is used inside anonymous functions. The caller may specify a signature.
//
// Deprecated: use github.com/fredbi/go-tracer/opencensus/tracer.StartNameSpan instead.
func StartNamedSpan(ctx context.Context, rt Loggable, signature string, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	return octracer.StartNamedSpan(ctx, rt, signature, fields...)
}
