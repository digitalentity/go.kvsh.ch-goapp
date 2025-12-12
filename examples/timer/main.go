package main

import (
	"context"
	"time"

	"go.kvsh.ch/goapp"
	"go.kvsh.ch/goapp/coremodules/timer"
	"go.kvsh.ch/goapp/module"
)

type Module struct {
	module.ModuleWithoutRun
}

func (m *Module) Name() module.Key {
	return module.Key("examples-timer")
}

func (m *Module) Depends() []module.Key {
	return []module.Key{
		timer.Key(),
	}
}

func (m *Module) TimerCallback(ctx context.Context, t time.Time) error {
	println("Timer fired at:", t.String())
	return nil
}

func (m *Module) Configure(b *module.Binder) error {
	t := b.Get(timer.Key()).(*timer.Module)
	t.Register(1*time.Second, m.TimerCallback)
	return nil
}

func New() *Module {
	return &Module{}
}

func main() {

	goapp.Install(New())
	goapp.Install(timer.New())

	goapp.Run(context.Background())
}
