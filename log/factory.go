package log

import (
	"context"

	"encoding/binary"

	octrace "go.opencensus.io/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Factory is wrapper for a logger.
	//
	// A factory wraps a logger to propagate log entries as trace spans.
	//
	// Factory supports both Opencensus and OpenTelemetry (OTEL) traces.
	// Use WithOTEL(true) option to build a factory with OTEL trace support.
	//
	// Loggers are zap structured loggers: see go.uber.org/zap.
	Factory struct {
		logger *zap.Logger
		fields []zap.Field

		options
	}
)

// NewFactory creates a new logger Factory for an underlying zap logger.
func NewFactory(logger *zap.Logger, opts ...Option) Factory {
	return Factory{
		logger:  logger.WithOptions(zap.AddCallerSkip(1)),
		options: defaultOptions(opts),
	}
}

// Bg creates a context-unaware logger, tied to the Background context.
func (b Factory) Bg() Logger {
	return logger{logger: b.logger}
}

// Zap returns the underlying zap logger
func (b Factory) Zap() *zap.Logger {
	return b.logger
}

// For returns a context-aware Logger.
//
// If the context contains a trace span, all logging calls are also
// echo-ed to that span.
//
// NOTE: for Datadog trace correlation, extra fields "dd.trace_id" and "dd.span_id"
// fields are added to the log entry.
func (b Factory) For(ctx context.Context) Logger {
	if b.otel {
		return b.forOTEL(ctx)
	}

	return b.forOpencensus(ctx)
}

func (b Factory) forOTEL(ctx context.Context) Logger {
	span := oteltrace.SpanFromContext(ctx)
	if span == nil {
		// OTEL tracing returns a noop span, not nil
		return b.Bg()
	}

	stx := span.SpanContext()
	logger := b.logger

	if b.datadog {
		// TODO: check DataDog OTEL exporter
		logger = loggerForDD(logger, stx.TraceID(), stx.SpanID())
	}

	return otelLogger{
		span:   span,
		fields: b.fields,
		logger: logger,
		ddFlag: b.datadog,
	}
}

func (b Factory) forOpencensus(ctx context.Context) Logger {
	span := octrace.FromContext(ctx)
	if span == nil {
		return b.Bg()
	}

	stx := span.SpanContext()
	logger := b.logger

	if b.datadog {
		logger = loggerForDD(logger, stx.TraceID, stx.SpanID)
	}

	return spanLogger{
		span:   span,
		fields: b.fields,
		logger: logger,
		ddFlag: b.datadog,
	}
}

// for datadog correlation, extract trace/span IDs as fields to add to the log entry.
//
// This corresponds to what the datadog opencensus exporter does:
// https://github.com/DataDog/opencensus-go-exporter-datadog/tree/master/span.go#L47
func loggerForDD(logger *zap.Logger, trID [16]byte, spID [8]byte) *zap.Logger {
	traceID := binary.BigEndian.Uint64(trID[8:])
	spanID := binary.BigEndian.Uint64(spID[:])
	return logger.With(
		zap.Uint64("dd.trace_id", traceID),
		zap.Uint64("dd.span_id", spanID),
		zap.Float64("sampling.priority", 1.00),
	)
}

// With creates a child Factory with some extra context fields.
func (b Factory) With(fields ...zapcore.Field) Factory {
	return Factory{
		logger:  b.logger.With(fields...),
		fields:  append(b.fields, fields...),
		options: b.options,
	}
}

// WithZapOptions creates a child Factory with some extra zap.Options for the underlying logger.
func (b Factory) WithZapOptions(opts ...zap.Option) Factory {
	return Factory{
		logger:  b.logger.WithOptions(opts...),
		fields:  b.fields,
		options: b.options,
	}
}
