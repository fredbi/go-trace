package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Option is an alias to go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp.Option
type Option = otelhttp.Option

// OTELHTTP is a middleware that wraps go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp.
//
// It allows a http server to initialize the trace context for handlers and to propagate incoming trace headers.
//
// Options available to otelhtpp.Handler are exposed here as Options.
func OTELHTTP(operation string, opts ...Option) func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware(operation, opts...)
}
