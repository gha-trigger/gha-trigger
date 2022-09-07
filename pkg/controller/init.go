package controller

import (
	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"go.uber.org/zap"
)

type Controller struct {
	cfg   *config.Config
	osEnv osenv.OSEnv
	ghs   map[int64]*githubapp.GitHubApp
}

func New(cfg *config.Config, logger *zap.Logger, osEnv osenv.OSEnv, ghs map[int64]*githubapp.GitHubApp) *Controller {
	return &Controller{
		cfg:   cfg,
		osEnv: osEnv,
		ghs:   ghs,
	}
}
