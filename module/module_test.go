package module

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ModuleA struct {
	ModuleWithoutRun
	configured bool
}

func (m *ModuleA) Name() Key {
	return "module-a"
}

func (m *ModuleA) Dependns() []Key {
	return []Key{"module-b"}
}

func (m *ModuleA) Configure(b *Binder) error {
	m.configured = true
	return nil
}

type ModuleB struct {
	ModuleWithoutRun
	ModuleWithoutDeps
	configured bool
}

func (m *ModuleB) Name() Key {
	return "module-b"
}

func (m *ModuleB) Configure(b *Binder) error {
	m.configured = true
	return nil
}

type ModuleC struct {
	ModuleWithoutRun
	configured bool
}

func (m *ModuleC) Name() Key {
	return "module-c"
}

func (m *ModuleC) Dependns() []Key {
	return []Key{"module-a", "module-b"}
}

func (m *ModuleC) Configure(b *Binder) error {
	m.configured = true
	return nil
}

// Modules D and E create a circular dependency for testing.
type ModuleD struct {
	ModuleWithoutRun
	ModuleWithoutConfigure
}

func (m *ModuleD) Name() Key {
	return "module-d"
}

func (m *ModuleD) Dependns() []Key {
	return []Key{"module-e"}
}

type ModuleE struct {
	ModuleWithoutRun
	ModuleWithoutConfigure
}

func (m *ModuleE) Name() Key {
	return "module-e"
}

func (m *ModuleE) Dependns() []Key {
	return []Key{"module-d"}
}

// Module F will call a Get() during Configure to test that behavior.
type ModuleF struct {
	ModuleWithoutRun
	configured bool
}

func (m *ModuleF) Name() Key {
	return "module-f"
}

func (m *ModuleF) Dependns() []Key {
	return []Key{"module-b"}
}

func (m *ModuleF) Configure(b *Binder) error {
	// Call Get on module-b to ensure it is configured first.
	_ = b.Get("module-b")
	m.configured = true
	return nil
}

// Module G will illegally call Get() in the Run() method to test that behavior.
type ModuleG struct {
	ModuleWithoutDeps // No dependencies, but will call Get illegally.
	binder            *Binder
	configured        bool
}

func (m *ModuleG) Name() Key {
	return "module-g"
}

func (m *ModuleG) Configure(b *Binder) error {
	m.binder = b
	m.configured = true
	return nil
}

func (m *ModuleG) Run(ctx context.Context) error {
	// Illegal: Call Get during Run.
	_ = m.binder.Get("module-b")
	return nil
}

// Module H returns an error during Configure.
type ModuleH struct {
	ModuleWithoutRun
	ModuleWithoutDeps
}

func (m *ModuleH) Name() Key {
	return "module-h"
}

func (m *ModuleH) Configure(b *Binder) error {
	return assert.AnError
}

func TestMultipleInstancesOfSameModule(t *testing.T) {
	b := NewBinder()
	modA1 := &ModuleA{}
	modA2 := &ModuleA{}

	assert.Panics(t, func() {
		b.Install(modA1)
		b.Install(modA2)
	})
}

func TestMissingDependency(t *testing.T) {
	b := NewBinder()
	b.Install(&ModuleA{})

	// This should panic, because module-a depends on module-b which is not installed.
	assert.Panics(t, func() {
		b.Get("module-a")
	})
}

func TestMultipleModules(t *testing.T) {
	b := NewBinder()
	modA := &ModuleA{}
	modB := &ModuleB{}
	modC := &ModuleC{}

	b.Install(modC)
	b.Install(modA)
	b.Install(modB)

	// Assert that we didn't panic and all modules are configured in the correct order.
	assert.NotPanics(t, func() {
		b.Get("module-c")
	})

	// C depends on A and B, so they should be configured first.
	assert.True(t, modB.configured, "module-b should be configured")
	assert.True(t, modA.configured, "module-a should be configured")
	assert.True(t, modC.configured, "module-c should be configured")
}

func TestCircularDependency(t *testing.T) {
	b := NewBinder()
	b.Install(&ModuleD{})
	b.Install(&ModuleE{})

	// This should panic due to circular dependency.
	assert.Panics(t, func() {
		b.Get("module-d")
	})
}

func TestRunModules(t *testing.T) {
	b := NewBinder()
	modA := &ModuleA{}
	modB := &ModuleB{}

	b.Install(modA)
	b.Install(modB)

	// This should not panic and run both modules.
	// Run implicitly configures the modules first.
	assert.NotPanics(t, func() {
		b.Run(context.Background())
	})

	assert.True(t, modA.configured, "module-a should be configured")
	assert.True(t, modB.configured, "module-b should be configured")
}

func TestGetDuringConfigure(t *testing.T) {
	b := NewBinder()
	modB := &ModuleB{}
	modF := &ModuleF{}

	b.Install(modB)
	b.Install(modF)

	// This should not panic. ModuleF calls Get on module-b during Configure.
	assert.NotPanics(t, func() {
		b.Get("module-f")
	})

	assert.True(t, modB.configured, "module-b should be configured")
	assert.True(t, modF.configured, "module-f should be configured")
}

func TestIllegalGetDuringRun(t *testing.T) {
	b := NewBinder()
	modB := &ModuleB{}
	modG := &ModuleG{}

	b.Install(modB)
	b.Install(modG)

	// This should panic. ModuleG calls Get on module-b during Run.
	assert.NotPanics(t, func() {
		b.Get("module-g")
	})

	// At this point modB should be configured, but modG should not have run yet.
	assert.True(t, modG.configured, "module-g should be configured")
	assert.False(t, modB.configured, "module-b should not be configured")

	// Run() will catch the panic and return it as an error.
	err := b.Run(context.Background())
	assert.Error(t, err, "expected error when running module-g due to illegal Get during Run")
}

func TestConfigureError(t *testing.T) {
	b := NewBinder()
	modH := &ModuleH{}

	b.Install(modH)

	// This should panic due to error during Configure.
	assert.Panics(t, func() {
		b.Get("module-h")
	})
}
