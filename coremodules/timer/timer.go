// Package timer provides a periodic ticker for scheduling tasks in the application.
package timer

import (
	"context"
	"fmt"
	"time"

	"go.kvsh.ch/goapp/module"

	"golang.org/x/sync/errgroup"
)

var Data = module.NewData[*Module]("timer")

type TimerCallbackFn func(ctx context.Context, t time.Time) error

type RegisteredTimer struct {
	Interval time.Duration
	Callback TimerCallbackFn
}

// This is a GoApp module that provides a periodic ticker.
type Module struct {
	module.ModuleWithoutDeps
	module.ModuleWithoutConfigure
	Timers []RegisteredTimer
}

func New() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return Data.Name()
}

func runTimerCallback(ctx context.Context, callback TimerCallbackFn, t time.Time) (err error) {
	// Catch panics from the callback and return them as errors.
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				e = fmt.Errorf("panic: %v", r)
			}
			err = fmt.Errorf("timer callback panicked: %w", e)
		}
	}()
	return callback(ctx, t)
}

func (m *Module) Run(ctx context.Context) error {

	eg, ectx := errgroup.WithContext(ctx)
	for _, t := range m.Timers {
		t := t
		eg.Go(func() error {
			ticker := time.NewTicker(t.Interval)
			defer ticker.Stop()
			for {
				select {
				case tickTime := <-ticker.C:
					if err := runTimerCallback(ectx, t.Callback, tickTime); err != nil {
						return err
					}
				case <-ectx.Done():
					return ectx.Err()
				}
			}
		})
	}
	return eg.Wait()
}

func (m *Module) Register(interval time.Duration, callback TimerCallbackFn) {
	m.Timers = append(m.Timers, RegisteredTimer{
		Interval: interval,
		Callback: callback,
	})
}
