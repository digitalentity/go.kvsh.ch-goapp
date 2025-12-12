package timer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New()
	assert.NotNil(t, m)
	assert.Empty(t, m.Timers)
}

func TestRegister(t *testing.T) {
	m := New()
	callback := func(ctx context.Context, t time.Time) error {
		return nil
	}
	m.Register(1*time.Second, callback)

	assert.Len(t, m.Timers, 1)
	assert.Equal(t, 1*time.Second, m.Timers[0].Interval)
}

func TestRun(t *testing.T) {
	m := New()

	called := make(chan bool, 1)
	callback := func(ctx context.Context, tm time.Time) error {
		called <- true
		return nil
	}
	m.Register(10*time.Millisecond, callback)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := m.Run(ctx); err != nil {
			assert.ErrorIs(t, err, context.Canceled)
		}
	}()

	select {
	case <-called:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Run() did not call the callback within the expected time")
	}
}

func TestRun_CallbackPanic(t *testing.T) {
	m := New()

	callback := func(ctx context.Context, tm time.Time) error {
		panic("test panic")
	}
	m.Register(10*time.Millisecond, callback)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.Run(ctx)
	}()

	select {
	case err := <-errCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timer callback panicked")
	case <-time.After(1 * time.Second):
		t.Fatal("Run() did not return an error within the expected time")
	}
}
