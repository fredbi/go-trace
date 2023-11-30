package log

import (
	"sync"

	"go.uber.org/zap/zapcore"
)

type fatalMock struct {
	mx    sync.Mutex
	count int
}

func (f *fatalMock) OnWrite(*zapcore.CheckedEntry, []zapcore.Field) {
	f.mx.Lock()
	f.count++
	f.mx.Unlock()
}
