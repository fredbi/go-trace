package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestSpanLogger(t *testing.T) {
	observed, observedLogs := observer.New(zapcore.DebugLevel)
	zlg, closer := MustGetLogger("root",
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, observed)
		})))
	defer closer()
	_, span := trace.StartSpan(context.Background(), "test")
	defer span.End()

	l := spanLogger{
		logger: zlg,
		span:   span,
	}
	ll := l.With(zap.String("string", "value"))
	require.NotNil(t, ll.Zap())

	t.Run("with logger fields", func(t *testing.T) {
		ll.Debug("debug")
		ll.Info("info")
		ll.Warn("warn")
		ll.Error("error")

		entries := observedLogs.All()
		require.Len(t, entries, 4)

		require.Equal(t, zapcore.DebugLevel, entries[0].Level)
		require.Equal(t, zapcore.InfoLevel, entries[1].Level)
		require.Equal(t, zapcore.WarnLevel, entries[2].Level)
		require.Equal(t, zapcore.ErrorLevel, entries[3].Level)

		for _, entry := range entries {
			require.NotEmpty(t, entry.Message)
			require.Len(t, entry.Context, 1)
		}
	})

	t.Run("with call-specific fields", func(t *testing.T) {
		ll.Info("more context", zap.Int("integer", 1))
		entries := observedLogs.All()
		require.Len(t, entries, 5)

		entry := entries[4]
		require.Equal(t, zapcore.InfoLevel, entry.Level)
		require.Len(t, entry.Context, 2)
	})

	t.Run("with Fatal level", func(t *testing.T) {
		fatalMock := &fatalMock{}
		zf, err := zap.NewProduction(zap.WithFatalHook(fatalMock))
		require.NoError(t, err)

		l := spanLogger{
			logger: zf,
		}

		l.Fatal("argh")

		require.Equal(t, 1, fatalMock.count)
	})
}
