package log

type (
	// Option to build a logger factory.
	Option func(*options)

	options struct {
		datadog bool
	}
)

var defaultFactoryOptions = options{}

func defaultOptions(opts []Option) options {
	if len(opts) == 0 {
		return defaultFactoryOptions
	}

	o := defaultFactoryOptions
	for _, apply := range opts {
		apply(&o)
	}

	return o
}

// WithDatadog enables datadog-specific correlation fields to link logs to trace spans
// (requires to export all trace samples).
func WithDatadog(enabled bool) Option {
	return func(o *options) {
		o.datadog = enabled
	}
}
