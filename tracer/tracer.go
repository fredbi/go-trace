package tracer

import (
	"context"
	"path"
	"runtime"

	"github.com/fredbi/go-trace/log"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

// Loggable is a log factory provider
type Loggable interface {
	Logger() log.Factory
}

var prefix = "function"

// StartSpan returns an opencensus span and logger that prepends the caller's signature.
//
// This spares us the boiler plate of repeatedly adding the prefix and function signatures in trace spans and logger.
func StartSpan(ctx context.Context, rt Loggable, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	pc, _, _, ok := runtime.Caller(1)
	var signature string
	if ok {
		signature = path.Base(runtime.FuncForPC(pc).Name())
	}

	sctx, span := trace.StartSpan(ctx, signature)

	signedFields := make([]zap.Field, 0, len(fields)+1)
	signedFields = append(signedFields, zap.String(prefix, signature))
	signedFields = append(signedFields, fields...)
	logger := rt.Logger().For(ctx).With(signedFields...)

	return sctx, span, logger
}

// StartNamedSpan is used inside anonymous functions. The caller may specify a signature.
func StartNamedSpan(ctx context.Context, rt Loggable, signature string, fields ...zap.Field) (context.Context, *trace.Span, log.Logger) {
	sctx, span := trace.StartSpan(ctx, signature)

	signedFields := make([]zap.Field, 0, len(fields)+1)
	signedFields = append(signedFields, zap.String(prefix, signature))
	signedFields = append(signedFields, fields...)
	logger := rt.Logger().For(ctx).With(signedFields...)

	return sctx, span, logger
}

// RegisterPrefix sets a package level prefix at initialization time.
//
// The default value is "function".
func RegisterPrefix(custom string) {
	prefix = custom
}
