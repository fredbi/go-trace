package log

import (
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
)

// Requests provides a middleware that logs http requests with a tracing-aware logger Factory.
func Requests(lf Factory) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lg := lf.For(r.Context())

			status := http.StatusOK
			rw = httpsnoop.Wrap(rw, httpsnoop.Hooks{
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					// will update the trace status when the chain of middlewares has completed
					return func(code int) {
						status = code
						next(code)
					}
				},
			})

			defer func() {
				lg.Info(
					"http request",
					zap.String("scheme", r.URL.Scheme),
					zap.Int("status", status),
					zap.String("method", r.Method),
					zap.String("uri", r.RequestURI),
					zap.Bool("tls", r.TLS != nil),
					zap.String("protocol", r.Proto),
					zap.String("host", r.Host),
					zap.String("remote_addr", r.RemoteAddr),
					zap.Duration("elapsed", time.Since(start)),
				)
			}()

			next.ServeHTTP(rw, r)
		})
	}
}
