package log

import (
	"encoding/base64"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// otelLogger copies the log output to an open telemetry (OTEL) trace span.
type otelLogger struct {
	logger *zap.Logger
	span   trace.Span
	fields []zap.Field
	ddFlag bool
}

func (sl otelLogger) Zap() *zap.Logger {
	return sl.logger
}

func (sl otelLogger) Debug(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.DebugLevel) {
		return
	}

	sl.logToSpan("debug", msg, fields...)
	sl.logger.Debug(msg, fields...)
}

func (sl otelLogger) Info(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.InfoLevel) {
		return
	}

	sl.logToSpan("info", msg, fields...)
	sl.logger.Info(msg, fields...)
}

func (sl otelLogger) Warn(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.WarnLevel) {
		return
	}

	sl.logToSpan("warn", msg, fields...)
	sl.logger.Warn(msg, fields...)
}

func (sl otelLogger) Error(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.ErrorLevel) {
		return
	}

	sl.logToSpan("error", msg, fields...)
	sl.logger.Error(msg, fields...)
}

func (sl otelLogger) Fatal(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.FatalLevel) {
		return
	}

	sl.logToSpan("fatal", msg, fields...)

	sl.logger.Fatal(msg, fields...)
}

// TODO: panic level

// With creates a child logger, with some extra context fields added to that logger.
func (sl otelLogger) With(fields ...zapcore.Field) Logger {
	return otelLogger{
		logger: sl.logger.With(fields...),
		fields: append(sl.fields, fields...),
		span:   sl.span,
		ddFlag: sl.ddFlag,
	}
}

func (sl otelLogger) logToSpan(level string, msg string, fields ...zapcore.Field) {
	fa := otelAdapter(make([]attribute.KeyValue, 0, len(sl.fields)+len(fields)+1))
	fa = append(fa, attribute.String("level", level))
	for _, field := range sl.fields {
		field.AddTo(&fa)
	}

	for _, field := range fields {
		field.AddTo(&fa)
	}

	sl.span.AddEvent(msg)
	sl.span.SetAttributes(fa...)

	/*
		if sl.ddFlag {
			// TODO: check that out when exporting to datadog, are events propagated?
			sl.span.SetAttributes(attribute.String("log_msg", msg))
		}
	*/
}

var _ zapcore.ObjectEncoder = &otelAdapter{}

// otelAdapter is a zapcore.ObjectEncoder to encode fields into a trace span.
//
// This instructs the zap logger to build the collection of trace attributes.
type otelAdapter []attribute.KeyValue

func (fa *otelAdapter) AddBool(key string, value bool) {
	*fa = append(*fa, attribute.Bool(key, value))
}

func (fa *otelAdapter) AddFloat64(key string, value float64) {
	*fa = append(*fa, attribute.Float64(key, value))
}

func (fa *otelAdapter) AddFloat32(key string, value float32) {
	*fa = append(*fa, attribute.Float64(key, float64(value)))
}

func (fa *otelAdapter) AddInt(key string, value int) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddInt64(key string, value int64) {
	*fa = append(*fa, attribute.Int64(key, value))
}

func (fa *otelAdapter) AddInt32(key string, value int32) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddInt16(key string, value int16) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddInt8(key string, value int8) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddUint(key string, value uint) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddUint64(key string, value uint64) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddUint32(key string, value uint32) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddUint16(key string, value uint16) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddUint8(key string, value uint8) {
	*fa = append(*fa, attribute.Int64(key, int64(value)))
}

func (fa *otelAdapter) AddDuration(key string, value time.Duration) {
	*fa = append(*fa, attribute.Stringer(key, value))
}

func (fa *otelAdapter) AddTime(key string, value time.Time) {
	*fa = append(*fa, attribute.Stringer(key, value))
}

func (fa *otelAdapter) AddBinary(key string, value []byte) {
	*fa = append(*fa, attribute.String(key, base64.StdEncoding.EncodeToString(value)))
}

func (fa *otelAdapter) AddByteString(key string, value []byte) {
	*fa = append(*fa, attribute.String(key, string(value)))
}

func (fa *otelAdapter) AddString(key, value string) {
	if key != "" && value != "" {
		*fa = append(*fa, attribute.String(key, value))
	}
}

// unsupported
func (fa *otelAdapter) AddUintptr(_ string, _ uintptr)                      {}
func (fa *otelAdapter) AddArray(_ string, _ zapcore.ArrayMarshaler) error   { return nil }
func (fa *otelAdapter) AddComplex128(_ string, _ complex128)                {}
func (fa *otelAdapter) AddComplex64(_ string, _ complex64)                  {}
func (fa *otelAdapter) AddObject(_ string, _ zapcore.ObjectMarshaler) error { return nil }
func (fa *otelAdapter) AddReflected(_ string, _ interface{}) error          { return nil }
func (fa *otelAdapter) OpenNamespace(_ string)                              {}
