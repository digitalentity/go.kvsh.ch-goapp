package goapp

import (
	"context"
	"os"
	"os/signal"

	"go.kvsh.ch/goapp/module"
)

func Run(ctx context.Context, modules ...module.Module) error {
	// Install all modules into the binder.
	b := module.NewBinder()
	for _, m := range modules {
		b.Install(m)
	}

	// Handler for signals
	sctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	return b.Run(sctx)
}
