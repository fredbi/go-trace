package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MustGetLogger creates a named zap logger, typically to inject into a service runtime as the root logger.
//
// This function returns a configured zap.Logger and a closing function to sync logs upon exit.
//
// It panics upon failures, such as invalid log level, or incapacity to build the underlying logger.
func MustGetLogger(name string, opts ...LoggerOption) (*zap.Logger, func()) {
	options := defaultLoggerOptions(opts)
	lc := zap.NewProductionConfig()
	if err := options.applyToConfig(&lc); err != nil {
		panic(fmt.Sprintf("parsing log level: %v", err))
	}

	zlg := zap.Must(
		lc.Build(zap.AddCallerSkip(options.callerSkip)),
	)
	zlg = zlg.Named(name)
	options.applyToLogger(zlg)

	return zlg, func() { _ = zlg.Sync() }
}

// GetTestLoggerConfig is intended to be used in test programs, and inject
// a logger factory or its underlying *zap.Logger into the tested components.
//
// It is configurable from the "DEBUG_TEST" environment variable: if set, logging
// is enabled. Otherwise, logging is just muted, allowing to keep test verbosity low.
func GetTestLoggerConfig(opts ...LoggerOption) (Factory, *zap.Logger, error) {
	isDebug := os.Getenv("DEBUG_TEST") != ""
	options := defaultLoggerOptions(opts)

	var zlg *zap.Logger
	if !isDebug {
		zlg = zap.NewNop()
		return NewFactory(zlg), zlg, nil
	}

	lc := zap.NewDevelopmentConfig()
	lc.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	lc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	lg, err := lc.Build(zap.AddCallerSkip(1))
	if err != nil {
		return Factory{}, nil, err
	}
	options.applyToLogger(zlg)
	zlg = lg

	factory := NewFactory(zlg)

	return factory, zlg, nil
}

// MustGetTestLoggerConfig is a wrapper around GetTestLoggerConfig that panics
// if an error is encountered.
func MustGetTestLoggerConfig() (Factory, *zap.Logger) {
	fl, zlg, err := GetTestLoggerConfig()
	if err != nil {
		panic(fmt.Sprintf("could not acquire test logger: %v", err))
	}

	return fl, zlg
}
