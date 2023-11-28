package tracer

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/fredbi/go-trace/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type mock struct {
	logger log.Factory
}

func (m mock) Logger() log.Factory {
	return m.logger
}

// testTracer implements a mock trace exporter
type testTracer struct {
	t        testing.TB
	exported []*trace.SpanData
	count    int
	mx       sync.Mutex
}

func (r *testTracer) ExportSpan(s *trace.SpanData) {
	b, _ := json.Marshal(s)
	r.t.Logf("%s", string(b))
	r.mx.Lock()
	r.count++
	r.exported = append(r.exported, s)
	r.mx.Unlock()
}

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
	myTestTracer := &testTracer{t: t}
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(myTestTracer)
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
		require.Equal(t, 1, myTestTracer.count)
		require.Len(t, myTestTracer.exported, 1)

		tr := myTestTracer.exported[0]

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

var mx sync.Mutex

func TestRegisterPrefix(t *testing.T) {
	const before = "function"
	t.Cleanup(func() {
		RegisterPrefix(before)
	})

	// don't want to pollute other tests
	mx.Lock()
	defer mx.Unlock()

	// ... but still want to demonstrate that the registration is goroutine-safe
	require.NotPanics(t, func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterPrefix("x")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterPrefix("y")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterPrefix("z")
		}()
	})
}

func testEnv(t testing.TB) (context.Context, Loggable) {
	t.Helper()

	zl, err := zap.NewDevelopment()
	require.NoError(t, err)

	zlf := log.NewFactory(zl, log.WithDatadog(true))

	rt := &mock{
		logger: zlf,
	}
	ctx := context.Background()

	return ctx, rt
}
