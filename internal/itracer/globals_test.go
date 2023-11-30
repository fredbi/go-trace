package itracer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterPrefix(t *testing.T) {
	const before = "function"
	t.Cleanup(func() {
		RegisterPrefix(before)
	})

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
