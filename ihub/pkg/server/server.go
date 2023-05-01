package server

import (
	"fmt"
	"ihub/pkg/config"
	"ihub/pkg/handler"
	"ihub/pkg/midware"

	"github.com/gin-gonic/gin"
)

type Server struct {
	r *gin.Engine
}

func (s *Server) Run() error {
	host := fmt.Sprintf(":%d", config.GetConfig().SERVER.Port)
	return s.r.Run(host)
}

func NewServer() *Server {
	s := &Server{}
	s.r = gin.New()
	//todo out cluster, get in/out config from yaml file
	//* InOut->Out->Auth->Approve->No->Endpoint
	//*      |                   |-> Yes -> Insert db
	//*      |-> In -> cluster gateway -> Auth -> Approve

	if err := midware.InitMidwares(s.r); err != nil {
		return nil
	}
	s.r.Use(midware.LoggerToFile(), midware.Auth(), gin.Recovery())
	s.r.GET("/health", handler.Health)
	s.r.Any("/:module/*endpoint", handler.Proxy)

	return s
}
