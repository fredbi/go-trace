package tracer_test

import (
	"context"

	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/tracer"
	"go.uber.org/zap"
)

var _ tracer.Loggable = Runtime{}

type Runtime struct {
	logger log.Factory
}

func (r Runtime) Logger() log.Factory {
	return r.logger
}

func ExampleStartSpan() {
	lg, _ := zap.NewProduction()
	rt := Runtime{logger: log.NewFactory(lg)}
	ctx := context.Background()

	// Instantiate a span and its associated span logger.
	//
	// This span is automatically signed with the current function, annotated with the source file and line
	_, span, logger := tracer.StartSpan(ctx, rt, zap.String("field", "fred"))
	defer span.End()

	logger.Info("test")

	// Should get something like:
	// 2023-11-01T17:19:58.615+0100	INFO	log/logger.go:35	test	{
	//	"function": "tracer.TestStartSpan",
	//	"source_file": ".../github.com/fredbi/go-trace/tracer/tracer_test.go",
	//	"source_line": 28,
	//	"field": "fred"
	//	}
}

func ExampleStartNamedSpan() {
	lg, _ := zap.NewDevelopment()
	rt := Runtime{logger: log.NewFactory(lg)}
	ctx := context.Background()

	// StartNamedSpan should be used in anonymous functions like so.
	handleFunc := func() {
		_, span, logger := tracer.StartNamedSpan(ctx, rt, "signature", zap.String("field", "fred"))
		defer span.End()

		logger.Info("test")
	}

	handleFunc()
}
