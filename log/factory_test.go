package log

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

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

func TestFactory(t *testing.T) {
	observed, observedLogs := observer.New(zapcore.DebugLevel)
	zlg, closer := MustGetLogger("root",
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, observed)
		})))
	defer closer()

	lgf := NewFactory(zlg) // builds a logger with trace propagation
	require.NotNil(t, lgf.Zap())

	t.Run("with background", func(t *testing.T) {
		l := lgf.Bg().With(zap.String("tescase", "bg test"))

		l.Warn("background log", zap.Int("entries", 1))

		entries := observedLogs.All()
		require.Len(t, entries, 1)

		entry := entries[0]

		require.Equal(t, zapcore.WarnLevel, entry.Level)
		require.Equal(t, "background log", entry.Message)
		require.NotEmpty(t, entry.Time)
		require.Contains(t, entry.Caller.String(), "factory_test.go:")
		require.Len(t, entry.Context, 2)
	})

	t.Run("with context, no span", func(t *testing.T) {
		type key string
		ctx := context.WithValue(context.Background(), key("x"), "y")
		l := lgf.For(ctx).With(zap.String("tescase", "bg test"))

		l.Warn("no span log", zap.Int("entries", 1))

		entries := observedLogs.All()
		require.Len(t, entries, 2)

		entry := entries[1]

		require.Equal(t, zapcore.WarnLevel, entry.Level)
		require.Equal(t, "no span log", entry.Message)
		require.NotEmpty(t, entry.Time)
		require.Contains(t, entry.Caller.String(), "factory_test.go:")
		require.Len(t, entry.Context, 2)
	})

	t.Run("with context span", func(t *testing.T) {
		myTestTracer := &testTracer{t: t}
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		trace.RegisterExporter(myTestTracer)
		t.Cleanup(func() {
			trace.UnregisterExporter(myTestTracer)
		})

		ctx, span := trace.StartSpan(context.Background(), "span name")
		l := lgf.For(ctx).With(zap.String("tescase", "bg test"))

		l.Info("span log", zap.Int("entries", 1))

		t.Run("assert log entry", func(t *testing.T) {
			entries := observedLogs.All()
			require.Len(t, entries, 3)

			entry := entries[2]

			require.Equal(t, zapcore.InfoLevel, entry.Level)
			require.Equal(t, "span log", entry.Message)
			require.NotEmpty(t, entry.Time)
			require.Contains(t, entry.Caller.String(), "factory_test.go:")
			require.Len(t, entry.Context, 2)
		})

		defer func() {
			t.Run("assert span entry", func(t *testing.T) {
				require.Equal(t, 1, myTestTracer.count)
				require.Len(t, myTestTracer.exported, 1)

				tr := myTestTracer.exported[0]

				require.NotEmpty(t, tr.SpanContext.TraceID)
				require.NotEmpty(t, tr.SpanContext.SpanID)

				require.NotEmpty(t, tr.StartTime)
				require.NotEmpty(t, tr.EndTime)

				require.Len(t, tr.Attributes, 3)
				require.Len(t, tr.Annotations, 1)
				require.Equal(t, "span log", tr.Annotations[0].Message)
			})
		}()
		defer span.End()
	})
}

func TestFactoryWithDatadog(t *testing.T) {
	observed, observedLogs := observer.New(zapcore.DebugLevel)
	zlg, closer := MustGetLogger("root",
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, observed)
		})))
	defer closer()

	lgf := NewFactory(zlg, WithDatadog(true))

	t.Run("with context span", func(t *testing.T) {
		myTestTracer := &testTracer{t: t}
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		trace.RegisterExporter(myTestTracer)
		t.Cleanup(func() {
			trace.UnregisterExporter(myTestTracer)
		})

		ctx, span := trace.StartSpan(context.Background(), "span name")
		l := lgf.For(ctx).With(zap.String("tescase", "bg test"))

		t.Run("internal verification", func(t *testing.T) {
			asSpanLogger, ok := l.(spanLogger)
			assert.True(t, ok)
			assert.True(t, asSpanLogger.ddFlag)
		})

		l.Info("span log", zap.Int("entries", 1))

		t.Run("assert log entry", func(t *testing.T) {
			entries := observedLogs.All()
			require.Len(t, entries, 1)

			entry := entries[0]

			require.Equal(t, zapcore.InfoLevel, entry.Level)
			require.Equal(t, "span log", entry.Message)
			require.NotEmpty(t, entry.Time)
			require.Contains(t, entry.Caller.String(), "factory_test.go:")
			require.Len(t, entry.Context, 5)

			fieldsFound := 0
			for _, field := range entry.Context {
				switch field.Key {
				case "dd.trace_id", "dd.span_id", "sampling.priority":
					fieldsFound++
					require.NotEmpty(t, field.Integer)
				}
			}
			require.Equal(t, 3, fieldsFound)
		})

		defer func() {
			t.Run("assert span entry", func(t *testing.T) {
				require.Equal(t, 1, myTestTracer.count)
				require.Len(t, myTestTracer.exported, 1)

				tr := myTestTracer.exported[0]

				require.NotEmpty(t, tr.SpanContext.TraceID)
				require.NotEmpty(t, tr.SpanContext.SpanID)

				require.NotEmpty(t, tr.StartTime)
				require.NotEmpty(t, tr.EndTime)

				require.Len(t, tr.Attributes, 4)
				require.Len(t, tr.Annotations, 1)
				require.Equal(t, "span log", tr.Annotations[0].Message)
				require.Contains(t, tr.Attributes, "log_msg")
			})
		}()
		defer span.End()
	})
}
