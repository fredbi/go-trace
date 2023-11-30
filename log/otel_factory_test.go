package log

import (
	"context"
	"testing"

	"github.com/fredbi/go-trace/otel/exporters/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFactoryWithOTEL(t *testing.T) {
	observed, observedLogs := observer.New(zapcore.DebugLevel)
	zlg, closer := MustGetLogger("root",
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, observed)
		})))
	defer closer()

	lgf := NewFactory(zlg, WithOTEL(true)) // builds a logger with trace propagation on OTEL
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
		myTestTracer := mock.NewTracer(t)
		myExporter := myTestTracer.Exporter()

		ctx, span := myTestTracer.Start(context.Background(), "span name")
		l := lgf.For(ctx).With(zap.String("tescase", "span test"))

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
				require.Equal(t, 1, myExporter.Count())
				require.Len(t, myExporter.Exported, 1)

				tr := myExporter.Exported[0]

				require.NotEmpty(t, tr.SpanContext().TraceID())
				require.NotEmpty(t, tr.SpanContext().SpanID())

				require.NotEmpty(t, tr.StartTime)
				require.NotEmpty(t, tr.EndTime)

				require.Len(t, tr.Attributes(), 3)
				events := tr.Events()
				require.Len(t, events, 1)
				event := events[0]
				require.Equal(t, "span log", event.Name)
			})
		}()
		defer func() {
			span.End()
			_ = myTestTracer.ForceFlush(ctx)
		}()
	})
}

func TestFactoryWithOTELDatadog(t *testing.T) {
	observed, observedLogs := observer.New(zapcore.DebugLevel)
	zlg, closer := MustGetLogger("root",
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, observed)
		})))
	defer closer()

	lgf := NewFactory(zlg, WithOTEL(true), WithDatadog(true))

	t.Run("with context span", func(t *testing.T) {
		myTestTracer := mock.NewTracer(t)
		myExporter := myTestTracer.Exporter()

		ctx, span := myTestTracer.Start(context.Background(), "span name")
		l := lgf.For(ctx).With(zap.String("tescase", "bg test"))

		t.Run("internal verification", func(t *testing.T) {
			asSpanLogger, ok := l.(otelLogger)
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
				require.Equal(t, 1, myExporter.Count())
				require.Len(t, myExporter.Exported, 1)

				tr := myExporter.Exported[0]

				require.NotEmpty(t, tr.SpanContext().TraceID())
				require.NotEmpty(t, tr.SpanContext().SpanID())

				require.NotEmpty(t, tr.StartTime)
				require.NotEmpty(t, tr.EndTime)

				require.Len(t, tr.Attributes(), 3)
				events := tr.Events()
				require.Len(t, events, 1)

				event := events[0]
				require.Equal(t, "span log", event.Name)
				require.Empty(t, event.Attributes)
			})
		}()

		defer func() {
			span.End()
			_ = myTestTracer.ForceFlush(ctx)
		}()
	})
}
