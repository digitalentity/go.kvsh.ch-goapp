package module

import "context"

type Key string

type Module interface {
	// Name returns the name of the module.which also serves as its unique identifier.
	Name() Key

	// Declares returns the module's dependencies and provisions.
	Depends() []Key

	// Configure sets up the module before it is run.
	// This needs to be deterministic, fast and not require a context.
	// Configure() is called automatically and in order based on dependencies.
	Configure(b *Binder) error

	// Run starts the module's main functionality. GoApp will exit when all modules have stopped.
	// This function should block until the module is no longer doing any work.
	Run(ctx context.Context) error
}

type ModuleWithoutDeps struct{}

func (m *ModuleWithoutDeps) Depends() []Key {
	return []Key{}
}

type ModuleWithoutConfigure struct{}

func (m *ModuleWithoutConfigure) Configure(b *Binder) error {
	return nil
}

type ModuleWithoutRun struct{}

func (m *ModuleWithoutRun) Run(ctx context.Context) error {
	return nil
}
