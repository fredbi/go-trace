package tracer

import (
	"context"
	"testing"

	"github.com/fredbi/go-trace/log"
	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type mock struct {
	logger log.Factory
}

func (m mock) Logger() log.Factory {
	return m.logger
}

func TestStartSpan(t *testing.T) {
	zl, _ := zap.NewDevelopment()
	rt := &mock{
		logger: log.NewFactory(zl),
	}
	ctx := context.Background()

	tctx, span, logger := StartSpan(ctx, rt, zap.String("field", "fred"))
	defer span.End()

	logger.Info("test")

	sctx := trace.FromContext(tctx).SpanContext()

	assert.NotEmpty(t, sctx.TraceID)
	assert.NotEmpty(t, sctx.SpanID)
}
