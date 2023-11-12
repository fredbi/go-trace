package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Logger is a simplified abstraction of the zap.Logger
	Logger interface {
		Debug(msg string, fields ...zapcore.Field)
		Info(msg string, fields ...zapcore.Field)
		Warn(msg string, fields ...zapcore.Field)
		Error(msg string, fields ...zapcore.Field)
		Fatal(msg string, fields ...zapcore.Field)
		With(fields ...zapcore.Field) Logger
		Zap() *zap.Logger
	}

	// logger delegates all calls to the underlying zap.Logger
	logger struct {
		logger *zap.Logger
	}
)

// Zap returns the underlying zap logger
func (l logger) Zap() *zap.Logger {
	return l.logger
}

// Debug logs an debug msg with fields
func (l logger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info msg with fields
func (l logger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs an warn msg with fields
func (l logger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error msg with fields
func (l logger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func (l logger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l logger) With(fields ...zapcore.Field) Logger {
	return logger{logger: l.logger.With(fields...)}
}
