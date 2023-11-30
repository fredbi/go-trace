package middleware

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// OCHTTP is a middleware that wraps go.opencensus.io/plugin/ochttp for a more idiomatic usage.
//
// It allows a http server to initialize the trace context for handlers and to propagate incoming trace headers.
//
// Options available to ochtpp.Handler are exposed here as Options.
func OCHTTP(opts ...OCHTTPOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		traceHandler := &ochttp.Handler{
			Handler: next,
		}

		if len(opts) == 0 {
			return traceHandler
		}

		o := applyOCHTTPOptions(opts)
		traceHandler.IsPublicEndpoint = o.isPublicEndpoint
		traceHandler.IsHealthEndpoint = o.healthEndpointFunc
		traceHandler.FormatSpanName = o.spanNameFunc

		if len(o.startOptions) > 0 {
			to := trace.StartOptions{}
			for _, apply := range o.startOptions {
				apply(&to)
			}
			traceHandler.StartOptions = to
		}
		if o.getStartOptions != nil {
			traceHandler.GetStartOptions = o.getStartOptions
		}
		if o.format != nil {
			traceHandler.Propagation = o.format
		}
		if o.withRoute != "" {
			return ochttp.WithRouteTag(traceHandler, o.withRoute)
		}
		return traceHandler
	}
}

type (
	OCHTTPOption func(*ocHTTPOptions)

	ocHTTPOptions struct {
		withRoute          string
		format             propagation.HTTPFormat
		getStartOptions    func(*http.Request) trace.StartOptions
		healthEndpointFunc func(*http.Request) bool
		spanNameFunc       func(*http.Request) string
		startOptions       []trace.StartOption
		isPublicEndpoint   bool
	}
)

func applyOCHTTPOptions(opts []OCHTTPOption) ocHTTPOptions {
	var o ocHTTPOptions

	for _, apply := range opts {
		apply(&o)
	}

	return o
}

// WithFormat overrides the trace propagation format (default is B3).
func WithFormat(format propagation.HTTPFormat) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.format = format
	}
}

// WithRoute decorates the trace with an extra route tag
func WithRoute(route string) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.withRoute = route
	}
}

// IsPublicEndpoint instructs the tracer to consider incoming trace metadata as
// a linked trace rather than a parent (for publicly accessible servers).
func IsPublicEndpoint(enabled bool) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.isPublicEndpoint = enabled
	}
}

// IsHealthEndpoint instructs the tracer to skip tracing when the function returns true.
// a linked trace rather than a parent (for publicly accessible servers).
//
// By default, paths like /healthz or /_ah/health are filtered out from tracing.
func IsHealthEndpoint(filter func(*http.Request) bool) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.healthEndpointFunc = filter
	}
}

// WithStartOptions enables trace start options such a overriding the span kind or
// adding a sampler. See go.opencensus.io/trace.StartOptions.
func WithStartOptions(opts ...trace.StartOption) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.startOptions = opts
	}
}

// WithGetStartOptions enables trace start options such a overriding the span kind or
// adding a sampler, on a per request basis.
func WithGetStartOptions(fn func(*http.Request) trace.StartOptions) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.getStartOptions = fn
	}
}

// WithFormatSpanName injects a function to determine the span name according to the request.
func WithFormatSpanName(fn func(*http.Request) string) OCHTTPOption {
	return func(o *ocHTTPOptions) {
		o.spanNameFunc = fn
	}
}
