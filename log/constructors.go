package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// LoggerOption sets options to tune the behavior of a logger.
	LoggerOption func(*loggerOptions)

	loggerOptions struct {
		logLevel          string
		callerSkip        int
		isDevelopment     bool
		disableStackTrace bool
		redirectStdLog    bool
		replaceGlobals    bool
	}
)

var loggerDefaults = loggerOptions{
	logLevel:   "info",
	callerSkip: 1,
}

func defaultLoggerOptions(opts []LoggerOption) loggerOptions {
	if len(opts) == 0 {
		return loggerDefaults
	}

	o := loggerDefaults
	for _, apply := range opts {
		apply(&o)
	}

	return o
}

// WithLevel sets the log level.
//
// The level value must parse to a valid zapcore.Level, i.e. one of error, warn, info, debug values.
// The default level is "info".
func WithLevel(level string) LoggerOption {
	return func(o *loggerOptions) {
		o.logLevel = level
	}
}

// WithDisableStackTrace disable stack printing for this logger
func WithDisableStackTrace(disabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.disableStackTrace = disabled
	}
}

// WithCallerSkip sets the number of frames in the stack to skip.
//
// By default, this is set to 1, so the logging function itself is skipped.
func WithCallerSkip(skipped int) LoggerOption {
	return func(o *loggerOptions) {
		o.callerSkip = skipped
	}
}

func WithRedirectStdLog(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.redirectStdLog = enabled
	}
}

func WithReplaceGlobals(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.replaceGlobals = enabled
	}
}

// MustGetLogger creates a named zap logger, typically to inject into a service runtime as the root logger.
//
// This function returns a configured zap.Logger and a closing function to sync logs upon exit.
//
// It panics upon failures, such as invalid log level, or incapacity to build the underlying logger.
func MustGetLogger(name string, opts ...LoggerOption) (*zap.Logger, func()) {
	options := defaultLoggerOptions(opts)

	lc := zap.NewProductionConfig()
	if options.isDevelopment {
		lc.Development = true
	}

	lvl, err := zapcore.ParseLevel(options.logLevel)
	if err != nil {
		panic(fmt.Sprintf("parsing log level: %v", err))
	}

	lc.Level = zap.NewAtomicLevelAt(lvl)
	lc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	if options.disableStackTrace {
		lc.DisableStacktrace = options.disableStackTrace
	}

	zlg := zap.Must(lc.Build(zap.AddCallerSkip(options.callerSkip)))
	zlg = zlg.Named(name)
	if options.replaceGlobals {
		zap.ReplaceGlobals(zlg)
	}

	if options.redirectStdLog {
		zap.RedirectStdLog(zlg)
	}

	return zlg, func() {
		_ = zlg.Sync()
	}
}
