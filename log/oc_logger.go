package log

import (
	"encoding/base64"
	"math"
	"time"

	"go.opencensus.io/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// spanLogger copies the log output to an opencensus trace span.
type spanLogger struct {
	logger *zap.Logger
	span   *trace.Span
	fields []zap.Field
	ddFlag bool
}

func (sl spanLogger) Zap() *zap.Logger {
	return sl.logger
}

func (sl spanLogger) Debug(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.DebugLevel) {
		return
	}

	sl.logToSpan("debug", msg, fields...)
	sl.logger.Debug(msg, fields...)
}

func (sl spanLogger) Info(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.InfoLevel) {
		return
	}

	sl.logToSpan("info", msg, fields...)
	sl.logger.Info(msg, fields...)
}

func (sl spanLogger) Warn(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.WarnLevel) {
		return
	}

	sl.logToSpan("warn", msg, fields...)
	sl.logger.Warn(msg, fields...)
}

func (sl spanLogger) Error(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.ErrorLevel) {
		return
	}

	sl.logToSpan("error", msg, fields...)
	sl.logger.Error(msg, fields...)
}

func (sl spanLogger) Fatal(msg string, fields ...zapcore.Field) {
	if !sl.logger.Core().Enabled(zapcore.FatalLevel) {
		return
	}

	sl.logToSpan("fatal", msg, fields...)

	sl.logger.Fatal(msg, fields...)
}

// TODO: panic level

// With creates a child logger, with some extra context fields added to that logger.
func (sl spanLogger) With(fields ...zapcore.Field) Logger {
	return spanLogger{
		logger: sl.logger.With(fields...),
		fields: append(sl.fields, fields...),
		span:   sl.span,
		ddFlag: sl.ddFlag,
	}
}

func (sl spanLogger) logToSpan(level string, msg string, fields ...zapcore.Field) {
	fa := fieldAdapter(make([]trace.Attribute, 0, len(sl.fields)+len(fields)+1))
	fa = append(fa, trace.StringAttribute("level", level))
	for _, field := range sl.fields {
		field.AddTo(&fa)
	}

	for _, field := range fields {
		field.AddTo(&fa)
	}

	sl.span.Annotate(nil, msg)
	sl.span.AddAttributes(fa...)

	if sl.ddFlag {
		// when exporting to datadog, annotations are lost: only attributes are propagated.
		sl.span.AddAttributes(trace.StringAttribute("log_msg", msg))
	}
}

var _ zapcore.ObjectEncoder = &fieldAdapter{}

// fieldAdapter is a zapcore.ObjectEncoder to encode fields into a trace span.
//
// This instructs the zap logger to build the collection of trace attributes.
type fieldAdapter []trace.Attribute

func (fa *fieldAdapter) AddBool(key string, value bool) {
	*fa = append(*fa, trace.BoolAttribute(key, value))
}

func (fa *fieldAdapter) AddFloat64(key string, value float64) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(math.Float64bits(value))))
}

func (fa *fieldAdapter) AddFloat32(key string, value float32) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(math.Float64bits(float64(value)))))
}

func (fa *fieldAdapter) AddInt(key string, value int) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddInt64(key string, value int64) {
	*fa = append(*fa, trace.Int64Attribute(key, value))
}

func (fa *fieldAdapter) AddInt32(key string, value int32) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddInt16(key string, value int16) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddInt8(key string, value int8) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddUint(key string, value uint) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddUint64(key string, value uint64) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddUint32(key string, value uint32) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddUint16(key string, value uint16) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddUint8(key string, value uint8) {
	*fa = append(*fa, trace.Int64Attribute(key, int64(value)))
}

func (fa *fieldAdapter) AddDuration(key string, value time.Duration) {
	*fa = append(*fa, trace.StringAttribute(key, value.String()))
}

func (fa *fieldAdapter) AddTime(key string, value time.Time) {
	*fa = append(*fa, trace.StringAttribute(key, value.String()))
}

func (fa *fieldAdapter) AddBinary(key string, value []byte) {
	*fa = append(*fa, trace.StringAttribute(key, base64.StdEncoding.EncodeToString(value)))
}

func (fa *fieldAdapter) AddByteString(key string, value []byte) {
	*fa = append(*fa, trace.StringAttribute(key, string(value)))
}

func (fa *fieldAdapter) AddString(key, value string) {
	if key != "" && value != "" {
		*fa = append(*fa, trace.StringAttribute(key, value))
	}
}

// unsupported
func (fa *fieldAdapter) AddUintptr(_ string, _ uintptr)                      {}
func (fa *fieldAdapter) AddArray(_ string, _ zapcore.ArrayMarshaler) error   { return nil }
func (fa *fieldAdapter) AddComplex128(_ string, _ complex128)                {}
func (fa *fieldAdapter) AddComplex64(_ string, _ complex64)                  {}
func (fa *fieldAdapter) AddObject(_ string, _ zapcore.ObjectMarshaler) error { return nil }
func (fa *fieldAdapter) AddReflected(_ string, _ interface{}) error          { return nil }
func (fa *fieldAdapter) OpenNamespace(_ string)                              {}
