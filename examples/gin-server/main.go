package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.kvsh.ch/goapp"

	"go.kvsh.ch/goapp/coremodules/ginhttp"
	"go.kvsh.ch/goapp/module"
)

type Server struct {
	module.ModuleWithoutRun
}

func (s *Server) Name() module.Key {
	return module.Key("demo-gin-server")
}

func (s *Server) Depends() []module.Key {
	return []module.Key{
		ginhttp.Key(),
	}
}

func (s *Server) HandlePing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (s *Server) Configure(b *module.Binder) error {
	g := b.Get(ginhttp.Key()).(*ginhttp.Module)
	g.GET("/ping", s.HandlePing)
	return nil
}

func New() *Server {
	return &Server{}
}

func main() {
	bundle := module.Bundle{}

	bundle.Add(New())
	bundle.Add(ginhttp.New(&ginhttp.Config{
		Address: ":8080",
	}))

	goapp.InstallBundle(bundle)

	goapp.Run(context.Background())
}
