package module

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

type bindingState int

const (
	bindingStateUnconfigured bindingState = iota
	bindingStateConfiguring
	bindingStateConfigured
)

type binding struct {
	state  bindingState
	module Module
}

type Binder struct {
	done    bool
	modules map[Key]*binding
}

func NewBinder() *Binder {
	return &Binder{
		done:    false,
		modules: make(map[Key]*binding),
	}
}

func (b *Binder) assureNotDone() {
	if b.done {
		panic("binder: already configured")
	}
}

func (b *Binder) Install(m Module) {
	b.assureNotDone()
	if _, exists := b.modules[m.Name()]; exists {
		panic("binder: module already installed: " + string(m.Name()))
	}

	b.modules[m.Name()] = &binding{
		state:  bindingStateUnconfigured,
		module: m,
	}

	slog.Info("binder: installed module", "name", m.Name())
}

// Get retrieves a module by its key.
// Get also ensures that the module and all its dependencies have been configured.
func (b *Binder) Get(key Key) Module {
	b.assureNotDone()
	return b.configureAndGetModule(key)
}

func (b *Binder) configureAndGetModule(key Key) Module {
	binding, exists := b.modules[key]
	if !exists {
		panic("binder: module not found: " + string(key))
	}

	switch binding.state {
	case bindingStateConfigured:
		return binding.module

	case bindingStateConfiguring:
		panic("binder: circular dependency detected for module: " + string(key))

	case bindingStateUnconfigured:
		// Mark as configuring to detect circular dependencies
		binding.state = bindingStateConfiguring

		// Configure dependencies first
		decl := binding.module.Depends()
		for _, depKey := range decl {
			b.configureAndGetModule(depKey)
		}

		// Now configure the module itself
		slog.Info("binder: configuring module", "name", binding.module.Name())
		err := binding.module.Configure(b)
		if err != nil {
			panic("binder: failed to configure module " + string(key) + ": " + err.Error())
		}

		// Mark as configured
		binding.state = bindingStateConfigured

		return binding.module
	default:
		panic("binder: unknown binding state for module: " + string(key))
	}
}

func runModule(m Module, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(error)
			if !ok {
				e = fmt.Errorf("panic: %v", r)
			}
			err = fmt.Errorf("module %s panicked: %w", m.Name(), e)
		}
	}()
	return m.Run(ctx)
}

func (b *Binder) Run(ctx context.Context) error {
	b.assureNotDone()

	// Collect all modules to run. This will implicitly configure them all.
	mods := make([]Module, 0, len(b.modules))
	for _, k := range maps.Keys(b.modules) {
		mods = append(mods, b.Get(k))
	}

	// Mark binder as done to prevent further modifications. Any calls to Install or Get will panic.
	b.done = true

	// Run all modules concurrently.
	eg, ectx := errgroup.WithContext(ctx)
	for _, m := range mods {
		mod := m
		eg.Go(func() error {
			return runModule(mod, ectx)
		})
	}

	return eg.Wait()
}
