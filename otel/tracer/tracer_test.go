package tracer

import (
	"context"
	"testing"
	"time"

	"github.com/fredbi/go-trace/log"
	"github.com/fredbi/go-trace/otel/exporters/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestStartSpan(t *testing.T) {
	ctx, rt := testEnv(t)

	tctx, span, logger := StartSpan(ctx, rt, zap.String("field", "fred"))
	defer span.End()

	logger.Info("test")

	sctx := trace.SpanFromContext(tctx).SpanContext()

	assert.NotEmpty(t, sctx.TraceID)
	assert.NotEmpty(t, sctx.SpanID)
}

func TestSpanContent(t *testing.T) {
	ctx, rt := testEnv(t)
	exporter := rt.tracer.Exporter()

	sctx, span, logger := StartSpan(ctx, rt,
		zap.String("field", "fred"),
		zap.Int("int", 1),
		zap.Duration("duration", time.Second),
	)
	logger.Info("test", zap.Bool("called", true))

	defer func() {
		require.Equal(t, 1, exporter.Count())
		require.Len(t, exporter.Exported, 1)

		tr := exporter.Exported[0]

		require.NotNil(t, tr.SpanContext())
		require.NotEmpty(t, tr.SpanContext().TraceID())
		require.NotEmpty(t, tr.SpanContext().SpanID())

		require.NotEmpty(t, tr.StartTime())
		require.NotEmpty(t, tr.EndTime())
		require.Equal(t, "tracer.TestSpanContent", tr.Name())

		require.Len(t, tr.Attributes(), 6)
		foundKeys := 0
		for _, attr := range tr.Attributes() {
			switch attr.Key {
			case "level", "function", "field", "int", "duration", "called":
				foundKeys++
			}
		}
		require.Equal(t, 6, foundKeys)

		events := tr.Events()
		require.Len(t, events, 1)
		annotations := events[0]
		require.Equal(t, "test", annotations.Name)
		require.Empty(t, annotations.Attributes)

	}()
	span.End()
	require.NoError(t, rt.tracer.ForceFlush(sctx))
}

func TestStartNamedSpan(t *testing.T) {
	ctx, rt := testEnv(t)

	tctx, span, logger := StartNamedSpan(ctx, rt, "anonymous", zap.String("field", "fred"))
	span.End()

	logger.Info("test")

	sctx := trace.SpanFromContext(tctx).SpanContext()

	assert.NotEmpty(t, sctx.TraceID())
	assert.NotEmpty(t, sctx.SpanID())
}

func testEnv(t testing.TB) (context.Context, *mockRuntime) {
	t.Helper()

	zl, err := zap.NewDevelopment()
	require.NoError(t, err)

	zlf := log.NewFactory(zl, log.WithDatadog(true), log.WithOTEL(true)) // TODO: refact
	rt := &mockRuntime{
		logger: zlf,
		tracer: mock.NewTracer(t),
	}
	ctx := context.Background()

	return ctx, rt
}
