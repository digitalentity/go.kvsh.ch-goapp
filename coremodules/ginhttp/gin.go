// Package ginhttp wraps the Gin router into a GoApp module for eaier dependency management
package ginhttp

import (
	"context"

	"github.com/gin-gonic/gin"

	"go.kvsh.ch/goapp/module"
)

var Data = module.NewData[*Module]("ginhttp")

type Config struct {
	Address string
}

type Module struct {
	module.ModuleWithoutDeps
	module.ModuleWithoutConfigure
	*gin.Engine

	c *Config
}

func New(cfg *Config) *Module {
	return &Module{
		Engine: gin.Default(),
		c:      cfg,
	}
}

func (m *Module) Name() string {
	return Data.Name()
}

func (m *Module) Run(ctx context.Context) error {
	// Gin Run() is a blocking call that does not support context cancellation.
	// We create our own scaffolding here to run the server and stop it gracefully.
	// We need to return an error if the server fails to start as well.

	srvErrChan := make(chan error, 1)

	go func() {
		srvErrChan <- m.Engine.Run(m.c.Address)
	}()

	select {
	case <-ctx.Done():
		// Context cancelled, shut down the server.
		return nil
	case err := <-srvErrChan:
		// Server encountered an error.
		return err
	}
}
