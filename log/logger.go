package log

import (
	"fmt"
	"os"

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
	}

	// logger delegates all calls to the underlying zap.Logger
	logger struct {
		logger *zap.Logger
	}
)

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

// TODO: reorganize

// MustGetTestLoggerConfig is a wrapper around GetTestLoggerConfig that panics
// if an error is encountered.
func MustGetTestLoggerConfig() (Factory, *zap.Logger) {
	fl, zlg, err := GetTestLoggerConfig()
	if err != nil {
		panic(fmt.Sprintf("could not acquire test logger: %v", err))
	}

	return fl, zlg

}

// GetTestLoggerConfig is intended to be used in test programs, and inject
// a logger factory or its underlying *zap.Logger into the tested components.
//
// It is configurable from the "DEBUG_TEST" environment variable: if set, logging
// is enabled. Otherwise, logging is just muted, allowing to keep test verbosity low.
func GetTestLoggerConfig() (Factory, *zap.Logger, error) {
	isDebug := os.Getenv("DEBUG_TEST") != ""

	var zlg *zap.Logger
	if !isDebug {
		zlg = zap.NewNop()
	} else {
		lc := zap.NewDevelopmentConfig()
		lc.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		lc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		lg, err := lc.Build(zap.AddCallerSkip(1))
		if err != nil {
			return Factory{}, nil, err
		}
		zap.RedirectStdLog(lg)

		zlg = lg
	}

	factory := NewFactory(zlg)

	return factory, zlg, nil
}
