package amplitude

import (
	api "github.com/renatoaf/amplitude-go/amplitude"
	"go.uber.org/zap"
)

type (
	// Option for the amplitude span exporter.
	Option func(*options)

	options struct {
		spanEncoder   SpanEncoder
		clientOptions *api.Options
		logger        *zap.Logger
		filters       Filters
	}

	// EncoderOption allows for some optional behavior of the DefaultSpanEncoder.
	EncoderOption  func(*encoderOptions)
	encoderOptions struct {
		app       string
		version   string
		eventType string
	}
)

func defaultOptions(opts ...Option) *options {
	o := &options{
		logger:      zap.NewNop(),
		spanEncoder: DefaultSpanEncoder(),
	}

	for _, apply := range opts {
		apply(o)
	}

	return o
}

func defaultEncoderOptions(opts ...EncoderOption) *encoderOptions {
	o := &encoderOptions{
		app:       "demo-amplitude",
		version:   "v0.0.1",
		eventType: "backend-event",
	}

	for _, apply := range opts {
		apply(o)
	}

	return o
}

// WithAmplitudeClientOptions sets the options for http API client for amplitude.
func WithAmplitudeClientOptions(clientOptions *api.Options) Option {
	return func(o *options) {
		o.clientOptions = clientOptions
	}
}

// WithSpanFilters sets the span filters for this exporter.
//
// By default, no filters are set.
func WithSpanFilters(filters ...Filter) Option {
	return func(o *options) {
		o.filters = filters
	}
}

// WithSpanEncoder sets the span encoder for this exporter.
func WithSpanEncoder(spanEncoder SpanEncoder) Option {
	return func(o *options) {
		o.spanEncoder = spanEncoder
	}
}

// WithLogger sets a logger for this exporter. Logged events are only errors
// that could pop up when operating the amplitude API.
//
// By default, the logger is a NOP (doesn't output anything).
func WithLogger(logger *zap.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
