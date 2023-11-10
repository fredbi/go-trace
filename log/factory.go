package log

import (
	"context"

	"encoding/binary"

	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type (
	// Factory is the default logging wrapper that can create
	// logger instances either contextless, or for a given context
	// (e.g. to propagate trace spans).
	//
	// A factory wraps a logger to propagate log entries as trace spans.
	//
	// Loggers are zap structured loggers: see go.uber.org/zap.
	Factory struct {
		logger *zap.Logger

		*options
	}
)

// NewFactory creates a new logger Factory for an underlying zap logger.
func NewFactory(logger *zap.Logger, opts ...Option) Factory {
	return Factory{
		logger:  logger,
		options: defaultOptions(opts),
	}
}

// Bg creates a context-unaware logger, tied to the Background context.
func (b Factory) Bg() Logger {
	return logger{logger: b.logger}
}

func (b Factory) Zap() *zap.Logger {
	return b.logger
}

// For returns a context-aware Logger.
//
// If the context contains an OpenTracing span, all logging calls are also
// echo-ed to that span.
//
// NOTE: for Datadog trace correlation, extra fields "dd.trace_id" and "dd.span_id"
// fields are added to the log entry.
func (b Factory) For(ctx context.Context) Logger {
	if span := trace.FromContext(ctx); span != nil { // TODO: support opentracing
		stx := span.SpanContext()
		logger := b.logger.WithOptions(zap.AddCallerSkip(1))

		if b.datadog {
			// for datadog correlation, extract trace/span IDs as fields to add to the log entry.
			// This corresponds to what the datadog opencensus exporter does:
			// https://github.com/DataDog/opencensus-go-exporter-datadog/tree/master/span.go#L47
			traceID := binary.BigEndian.Uint64(stx.TraceID[8:])
			spanID := binary.BigEndian.Uint64(stx.SpanID[:])
			logger = logger.With(
				zap.Uint64("dd.trace_id", traceID),
				zap.Uint64("dd.span_id", spanID),
				zap.Float64(ext.SamplingPriority, 1.00),
			)
		}

		return spanLogger{
			span:   span,
			logger: logger,
		}
	}
	return b.Bg()
}

// With creates a child logger and optionally adds some context fields to that logger.
func (b Factory) With(fields ...zapcore.Field) Factory {
	return Factory{logger: b.logger.With(fields...)}
}

// WithZapOptions creates a child logger with some zap.Option
func (b Factory) WithZapOptions(opts ...zap.Option) Factory {
	return Factory{logger: b.logger.WithOptions(opts...)}
}
