package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.kvsh.ch/goapp"

	"go.kvsh.ch/goapp/coremodules/ginhttp"
	"go.kvsh.ch/goapp/module"
)

var Data = module.NewData[*Server]("demo-gin-server")

type Server struct {
	module.ModuleWithoutRun
}

func (s *Server) Name() string {
	return Data.Name()
}

func (s *Server) Depends() []module.Key {
	return []module.Key{
		ginhttp.Data,
	}
}

func (s *Server) HandlePing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (s *Server) Configure(b *module.Binder) error {
	g := ginhttp.Data.Get(b)
	g.GET("/ping", s.HandlePing)
	return nil
}

func New() *Server {
	return &Server{}
}

func main() {

	goapp.Install(New())
	goapp.Install(ginhttp.New(&ginhttp.Config{
		Address: ":8080",
	}))

	goapp.Run(context.Background())
}
