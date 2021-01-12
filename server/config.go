package server

import (
	"os"

	"github.com/kraem/zhuyi-go/pkg/env"
	"github.com/kraem/zhuyi-go/pkg/log"
	"github.com/kraem/zhuyi-go/network"
)

const LISTEN_ADDR = "LISTEN_ADDR"

type Server struct {
	CfgNetwork *network.Config
	Cfg       *config
}

type config struct {
	Addr string
}

func NewServer() *Server {
	c, err := network.NewConfig()
	if err != nil {
		log.LogError(err)
		os.Exit(1)
	}
	return &Server{
		CfgNetwork: c,
		Cfg:       Config(),
	}
}

func Config() *config {
	return &config{
		Addr: env.GetEnvOrExit(LISTEN_ADDR),
	}
}
