package tracer

import (
	"testing"
	"time"

	"github.com/fredbi/go-trace/opencensus/exporters/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

func TestStartSpan(t *testing.T) {
	ctx, rt := testEnv(t)

	tctx, span, logger := StartSpan(ctx, rt, zap.String("field", "fred"))
	defer span.End()

	logger.Info("test")

	sctx := trace.FromContext(tctx).SpanContext()

	assert.NotEmpty(t, sctx.TraceID)
	assert.NotEmpty(t, sctx.SpanID)
}

func TestSpanContent(t *testing.T) {
	myTestTracer := mock.New(t)
	t.Cleanup(func() {
		trace.UnregisterExporter(myTestTracer)
	})

	ctx, rt := testEnv(t)

	_, span, logger := StartSpan(ctx, rt,
		zap.String("field", "fred"),
		zap.Int("int", 1),
		zap.Duration("duration", time.Second),
	)
	logger.Info("test", zap.Bool("called", true))

	defer func() {
		require.Equal(t, 1, myTestTracer.Count())
		require.Len(t, myTestTracer.Exported, 1)

		tr := myTestTracer.Exported[0]

		require.NotEmpty(t, tr.SpanContext.TraceID)
		require.NotEmpty(t, tr.SpanContext.SpanID)

		require.NotEmpty(t, tr.StartTime)
		require.NotEmpty(t, tr.EndTime)
		require.Equal(t, "tracer.TestSpanContent", tr.Name)

		require.Len(t, tr.Attributes, 7)
		require.Contains(t, tr.Attributes, "level")
		require.Contains(t, tr.Attributes, "function")
		require.Contains(t, tr.Attributes, "field")
		require.Contains(t, tr.Attributes, "int")
		require.Contains(t, tr.Attributes, "duration")
		require.Contains(t, tr.Attributes, "called")
		require.Contains(t, tr.Attributes, "log_msg")

		require.Len(t, tr.Annotations, 1)
		annotations := tr.Annotations[0]
		require.Equal(t, "test", annotations.Message)
		require.Empty(t, annotations.Attributes)
	}()
	defer span.End()
}

func TestStartNamedSpan(t *testing.T) {
	ctx, rt := testEnv(t)

	tctx, span, logger := StartNamedSpan(ctx, rt, "anonymous", zap.String("field", "fred"))
	defer span.End()

	logger.Info("test")

	sctx := trace.FromContext(tctx).SpanContext()

	assert.NotEmpty(t, sctx.TraceID)
	assert.NotEmpty(t, sctx.SpanID)
}
