package log_test

import (
	"context"
	"os"

	"github.com/fredbi/go-trace/log"
	"go.opencensus.io/trace"
)

func ExampleMustGetLogger() {
	const appName = "my_app"

	zlog, closer := log.MustGetLogger(
		appName,
		log.WithLevel("debug"),
		log.WithOutput(log.Stdout),
	)

	zlog.Debug("test")
	defer closer()
}

func ExampleNewFactory() {
	ctx := context.Background()
	zlg, closer := log.MustGetLogger("root") // builds a named zap logger with sensible defaults
	defer closer()

	lgf := log.NewFactory(zlg) // builds a logger with trace propagation

	ctx, span := trace.StartSpan(ctx, "span name")
	defer span.End()
	lg := lgf.For(ctx)

	lg.Info("log propagated as a trace span")
}

func ExampleMustGetTestLoggerConfig() {
	os.Setenv("DEBUG_TEST", "1")
	zlf, zlg := log.MustGetTestLoggerConfig()

	zlg.Info("this is a logger visible only when DEBUG_TEST is set, e.g. for local testing")
	zlf.Bg().Info("this is a logger factory only visble when DEBUG_TEST is set, e.g. for local testing")
}
