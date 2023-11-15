package log

import (
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
		ignoreLevelErr    bool
	}
)

var loggerDefaults = loggerOptions{
	logLevel:   "info",
	callerSkip: 0,
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

func (o loggerOptions) applyToConfig(lc *zap.Config) error {
	if o.isDevelopment {
		lc.Development = true
	}

	lvl, err := zapcore.ParseLevel(o.logLevel)
	if err != nil {
		if o.ignoreLevelErr {
			lvl, _ = zapcore.ParseLevel(loggerDefaults.logLevel)
		} else {
			return err
		}
	}

	lc.Level = zap.NewAtomicLevelAt(lvl)
	lc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	if o.disableStackTrace {
		lc.DisableStacktrace = o.disableStackTrace
	}

	return nil
}

func (o loggerOptions) applyToLogger(zlg *zap.Logger) {
	if o.replaceGlobals {
		zap.ReplaceGlobals(zlg)
	}

	if o.redirectStdLog {
		zap.RedirectStdLog(zlg)
	}
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

// WithIgnoreErr ignores errors when parsing an invalid log level,
// and silently substitutes the default instead.
func WithIgnoreErr(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.ignoreLevelErr = enabled
	}
}
