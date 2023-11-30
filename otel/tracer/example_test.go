package tracer_test

import (
	"context"
	stdlog "log"
	"sync"

	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/otel/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	_              tracer.Traceable = Runtime{}
	onceInitTracer sync.Once
	tp             *sdktrace.TracerProvider
)

// Runtime is Traceable
type Runtime struct {
	tracer trace.Tracer
	logger log.Factory
}

func (r Runtime) Logger() log.Factory {
	return r.logger
}
func (r Runtime) Tracer() trace.Tracer {
	return r.tracer
}

func newExampleTracer() (trace.Tracer, func(context.Context) error) {
	onceInitTracer.Do(func() {
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			stdlog.Fatalf("failed to initialize stdouttrace exporter: %v", err)

			return
		}

		bsp := sdktrace.NewBatchSpanProcessor(exp)
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(bsp),
		)
		otel.SetTracerProvider(tp)
	})

	return otel.Tracer("example_tracer"), tp.ForceFlush
}

func executionContext() (context.Context, Runtime, func(context.Context) error) {
	lg := zap.NewExample()
	tr, flusher := newExampleTracer()
	rt := Runtime{
		logger: log.NewFactory(lg, log.WithOTEL(true)),
		tracer: tr,
	}
	ctx := context.Background()

	return ctx, rt, flusher
}

func ExampleStartSpan() {
	ctx, rt, flusher := executionContext()

	// Instantiate a span and its associated span logger.
	//
	// This span is automatically signed with the current function, annotated with the source file and line
	spanCtx, span, logger := tracer.StartSpan(ctx, rt, zap.String("field", "fred"))
	defer func() {
		span.End()
		_ = flusher(spanCtx) // flush the spans, so we get them in the output of this example
	}()

	logger.Info("test")

	// The trace span goes to stderr

	// output:
	// {"level":"info","msg":"test","function":"tracer_test.ExampleStartSpan","field":"fred"}
}

func ExampleStartNamedSpan() {
	ctx, rt, flusher := executionContext()

	// StartNamedSpan should be used in anonymous functions like so.
	handleFunc := func() {
		spanCtx, span, logger := tracer.StartNamedSpan(ctx, rt, "signature", zap.String("field", "fred"))
		defer func() {
			span.End()
			_ = flusher(spanCtx) // flush the spans, so we get them in the output of this example
		}()

		logger.Info("test")
	}

	handleFunc()

	// The trace span goes to stderr

	// output:
	// {"level":"info","msg":"test","function":"signature","field":"fred"}
}
