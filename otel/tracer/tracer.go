package tracer

import (
	"context"

	"github.com/fredbi/go-trace/internal/itracer"
	"github.com/fredbi/go-trace/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Loggable provides a log factory.
type Loggable interface {
	Logger() log.Factory
}

// Traceable provides a log factory and an OTEL tracer
type Traceable interface {
	Logger() log.Factory
	Tracer() trace.Tracer
}

// StartSpan returns an opencensus span and logger that prepends the caller's signature.
//
// This spares the boiler plate of repeatedly adding the prefix and function signatures in trace spans and logger.
//
// If the passed runtime "rt" is Traceable (i.e. provides an OTEL tracer), it will be used.
// Otherwise, the globally registered tracer will be returned.
// StartSpan will panic if no such tracer has been registered.
func StartSpan(ctx context.Context, rt Loggable, fields ...zap.Field) (context.Context, trace.Span, log.Logger) {
	tracer := getTracer(rt)
	signature := itracer.Signature()
	sctx, span := tracer.Start(ctx, signature)
	logger := rt.Logger().With(itracer.SignedFields(signature, fields)...).For(sctx)

	return sctx, span, logger
}

// StartNamedSpan is used inside anonymous functions. The caller may specify a signature.
//
// If the passed runtime "rt" is Traceable (i.e. provides an OTEL tracer), it will be used.
// Otherwise, the globally registered tracer will be returned.
// StartNamedSpan will panic if no such tracer has been registered.
func StartNamedSpan(ctx context.Context, rt Loggable, signature string, fields ...zap.Field) (context.Context, trace.Span, log.Logger) {
	tracer := getTracer(rt)
	sctx, span := tracer.Start(ctx, signature)
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

func getTracer(rt Loggable) trace.Tracer {
	if traceable, ok := rt.(Traceable); ok {
		return traceable.Tracer()
	}

	return otel.Tracer("") // returns the global default tracer
}
