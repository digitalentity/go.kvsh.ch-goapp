package goapp

import (
	"context"
	"os"
	"os/signal"

	"go.kvsh.ch/goapp/module"
)

var defaultBinder = module.NewBinder()

func Install(m module.Module) {
	defaultBinder.Install(m)
}

func Run(ctx context.Context) error {
	sctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	return defaultBinder.Run(sctx)
}
