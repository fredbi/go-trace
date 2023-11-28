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
		output            string
		seedConfig        func() zap.Config
		seedTestConfig    func() zap.Config
		zapOpts           []zap.Option
		callerSkip        int
		isDevelopment     bool
		disableStackTrace bool
		redirectStdLog    bool
		replaceGlobals    bool
		ignoreLevelErr    bool
		withCaller        bool
	}
)

var loggerDefaults = loggerOptions{
	logLevel:       "info",
	callerSkip:     0,
	withCaller:     true,
	seedConfig:     zap.NewProductionConfig,
	seedTestConfig: zap.NewDevelopmentConfig,
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
	lc.DisableCaller = !o.withCaller

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
//
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

// WithCaller adds the source location of the log entry.
//
// This is enabled by default.
func WithCaller(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.withCaller = enabled
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

// WithRedirectStdLog redirects the global standard library "log" to this logger.
func WithRedirectStdLog(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.redirectStdLog = enabled
	}
}

// WithReplaceGlobals replaces zap global loggers to ths logger.
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

type Output string

const (
	Stdout Output = "stdout"
	Stderr Output = "stderr"
)

func (o Output) String() string {
	return string(o)
}

// WithOutput allows to select stdout or stderr as the log sink.
//
// The default is stderr.
func WithOutput(output Output) LoggerOption {
	return func(o *loggerOptions) {
		o.output = output.String()
	}
}

// WithZapOptions injects zap.Options to the logger.
func WithZapOptions(opts ...zap.Option) LoggerOption {
	return func(o *loggerOptions) {
		o.zapOpts = append(o.zapOpts, opts...)
	}
}

// WithDevelopment puts the logger in development mode, which changes the
// behavior of DPanicLevel and takes stacktraces more liberally.
func WithDevelopment(enabled bool) LoggerOption {
	return func(o *loggerOptions) {
		o.isDevelopment = enabled
	}
}

// WithSeedConfig provides the seed zap.Config to construct a logger.
//
// By default, this is zap.NewProductionConfig().
//
// This is useful to set advanced, non default settings such as a colored encoder,
// a log sampler, etc.
func WithSeedConfig(seeder func() zap.Config) LoggerOption {
	return func(o *loggerOptions) {
		o.seedConfig = seeder
	}
}

// WithSeedTestConfig provides the seed zap.Config to construct a test logger.
//
// The default is zap.NewDevelopmentConfig() (whenever not muted).
func WithSeedTestConfig(seeder func() zap.Config) LoggerOption {
	return func(o *loggerOptions) {
		o.seedTestConfig = seeder
	}
}
