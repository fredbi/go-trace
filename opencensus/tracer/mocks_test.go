package tracer

import (
	"context"
	"testing"

	"github.com/fredbi/go-trace/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockRuntime struct {
	logger log.Factory
}

func (m mockRuntime) Logger() log.Factory {
	return m.logger
}

func testEnv(t testing.TB) (context.Context, Loggable) {
	t.Helper()

	zl, err := zap.NewDevelopment()
	require.NoError(t, err)

	zlf := log.NewFactory(zl, log.WithDatadog(true))

	rt := &mockRuntime{
		logger: zlf,
	}
	ctx := context.Background()

	return ctx, rt
}
