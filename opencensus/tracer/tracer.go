package tracer

import (
	"context"

	"github.com/fredbi/go-trace/internal/itracer"
	"github.com/fredbi/go-trace/log"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

// Loggable is a log factory provider
type Loggable interface {
	Logger() log.Factory
}

// StartSpan returns an opencensus span and logger that prepends the caller's signature.
//
// NOTE: Opencensus trace supports only one global tracer.
//
// This spares the boiler plate of repeatedly adding the prefix and function signatures in trace spans and logger.
func StartSpan(ctx context.Context, rt Loggable, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	signature := itracer.Signature()
	sctx, span := trace.StartSpan(ctx, signature)
	logger := rt.Logger().With(itracer.SignedFields(signature, fields)...).For(sctx)

	return sctx, span, logger
}

// StartNamedSpan is used inside anonymous functions. The caller may specify a signature.
func StartNamedSpan(ctx context.Context, rt Loggable, signature string, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	sctx, span := trace.StartSpan(ctx, signature)
	logger := rt.Logger().With(itracer.SignedFields(signature, fields)...).For(sctx)

	return sctx, span, logger
}

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
	itracer.RegisterPrefix(custom)
}
