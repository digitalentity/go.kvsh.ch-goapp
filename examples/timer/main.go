package main

import (
	"context"
	"time"

	"go.kvsh.ch/goapp"
	"go.kvsh.ch/goapp/coremodules/timer"
	"go.kvsh.ch/goapp/module"
)

var Data = module.NewData[*Module]("demo-timer")

type Module struct {
	module.ModuleWithoutRun
}

func (m *Module) Name() string {
	return Data.Name()
}

func (m *Module) Depends() []module.Key {
	return []module.Key{
		timer.Data,
	}
}

func (m *Module) TimerCallback(ctx context.Context, t time.Time) error {
	println("Timer fired at:", t.String())
	return nil
}

func (m *Module) Configure(b *module.Binder) error {
	t := timer.Data.Get(b)
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
