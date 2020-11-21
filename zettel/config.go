package zettel

import (
	"github.com/kraem/zhuyi-go/pkg/env"
	"github.com/kraem/zhuyi-go/pkg/fs"
)

const ZETTEL_PATH = "ZETTEL_PATH"

type Config struct {
	ZettelPath string
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		ZettelPath: fs.AppendTrailingSlash(env.GetEnvOrExit(ZETTEL_PATH)),
	}
	if err := fs.HavePermissions(cfg.ZettelPath); err != nil {
		return nil, err
	}
	return cfg, nil
}
