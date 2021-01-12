package network

import (
	"github.com/kraem/zhuyi-go/pkg/env"
	"github.com/kraem/zhuyi-go/pkg/fs"
)

const NETWORK_PATH = "NETWORK_PATH"

type Config struct {
	NetworkPath string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		NetworkPath: fs.AppendTrailingSlash(env.GetEnvOrExit(NETWORK_PATH)),
	}
	if err := fs.HavePermissions(cfg.NetworkPath); err != nil {
		return nil, err
	}
	return cfg, nil
}
