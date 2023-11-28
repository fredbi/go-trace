package log

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestMustGetLogger(t *testing.T) {
	const appName = "my_app"
	observed, observedLogs := observer.New(zapcore.DebugLevel)

	zlog, closer := MustGetLogger(
		appName,
		WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			// forwards logged entries to an observable sink
			return zapcore.NewTee(c, observed)
		})),
	)
	defer closer()

	zlog.Info("test", zap.String("context", "x"))

	entries := observedLogs.All()
	require.Len(t, entries, 1)
	entry := entries[0]

	require.Equal(t, zapcore.InfoLevel, entry.Level)
	require.NotEmpty(t, entry.Time)
	require.Equal(t, appName, entry.LoggerName)
	require.Equal(t, "test", entry.Message)
	require.Len(t, entry.Context, 1)
}

var testMux sync.Mutex

func TestMustGetTestLogger(t *testing.T) {
	testMux.Lock()
	defer testMux.Unlock()

	t.Run("with no env", mustGetTestLogger(""))

	t.Run("with env", mustGetTestLogger("1"))
}

func mustGetTestLogger(env string) func(*testing.T) {
	return func(t *testing.T) {
		t.Setenv("DEBUG_TEST", env)

		observed, observedLogs := observer.New(zapcore.DebugLevel)

		zlf, zlg := MustGetTestLoggerConfig(
			WithZapOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewTee(c, observed)
			})),
		)
		require.NotNil(t, zlf)
		require.NotNil(t, zlg)

		l := zlf.Bg()
		l.Info("test")

		entries := observedLogs.All()
		if env == "" {
			require.Len(t, entries, 0)

			return
		}

		require.Len(t, entries, 1)
	}
}
