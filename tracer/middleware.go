package tracer

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// Middleware that wraps go.opencensus.io/plugin/ochttp for a more idiomatic usage.
//
// Options available to ochtpp.Handler are exposed here as Options.
func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		traceHandler := &ochttp.Handler{
			Handler: next,
		}

		if len(opts) == 0 {
			return traceHandler
		}

		o := applyMiddlewareOptions(opts)
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
	MiddlewareOption func(*middlewareOptions)

	middlewareOptions struct {
		withRoute          string
		format             propagation.HTTPFormat
		getStartOptions    func(*http.Request) trace.StartOptions
		healthEndpointFunc func(*http.Request) bool
		spanNameFunc       func(*http.Request) string
		startOptions       []trace.StartOption
		isPublicEndpoint   bool
	}
)

func applyMiddlewareOptions(opts []MiddlewareOption) middlewareOptions {
	var o middlewareOptions

	for _, apply := range opts {
		apply(&o)
	}

	return o
}

// WithFormat overrides the trace propagation format (default is B3).
func WithFormat(format propagation.HTTPFormat) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.format = format
	}
}

// WithRoute decorates the trace with an extra route tag
func WithRoute(route string) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.withRoute = route
	}
}

// IsPublicEndpoint instructs the tracer to consider incoming trace metadata as
// a linked trace rather than a parent (for publicly accessible servers).
func IsPublicEndpoint(enabled bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.isPublicEndpoint = enabled
	}
}

// IsHealthEndpoint instructs the tracer to skip tracing when the function returns true.
// a linked trace rather than a parent (for publicly accessible servers).
//
// By default, paths like /healthz or /_ah/health are filtered out from tracing.
func IsHealthEndpoint(filter func(*http.Request) bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.healthEndpointFunc = filter
	}
}

// WithStartOptions enables trace start options such a overriding the span kind or
// adding a sampler. See go.opencensus.io/trace.StartOptions.
func WithStartOptions(opts ...trace.StartOption) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.startOptions = opts
	}
}

// WithGetStartOptions enables trace start options such a overriding the span kind or
// adding a sampler, on a per request basis.
func WithGetStartOptions(fn func(*http.Request) trace.StartOptions) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.getStartOptions = fn
	}
}

// WithFormatSpanName injects a function to determine the span name according to the request.
func WithFormatSpanName(fn func(*http.Request) string) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.spanNameFunc = fn
	}
}
