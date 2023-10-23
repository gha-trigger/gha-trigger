package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/aws"
	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/controller"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Handler struct {
	logger *zap.Logger
	ctrl   Controller
}

type Controller interface {
	Do(ctx context.Context, logger *zap.Logger, req *domain.Request) error
}

func New(ctx context.Context, logger *zap.Logger) (*Handler, error) {
	// read config
	cfg := &config.Config{}
	osEnv := osenv.New()
	if err := readConfig(cfg, osEnv); err != nil {
		return nil, err
	}
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration is invalid: %w", err)
	}
	if err := config.Init(cfg); err != nil {
		return nil, fmt.Errorf("initialize configuration: %w", err)
	}
	// read secret
	awsClient := aws.New(cfg.AWS)
	numGitHubApps := len(cfg.GitHubApps)
	ghApps := make(map[int64]*githubapp.GitHubApp, numGitHubApps)
	ghs := make(map[string]*github.Client, numGitHubApps)
	for i := 0; i < numGitHubApps; i++ {
		appCfg := cfg.GitHubApps[i]
		ghApp, err := githubapp.New(ctx, awsClient, appCfg)
		if err != nil {
			return nil, err
		}
		ghApps[appCfg.AppID] = ghApp
		ghs[appCfg.Name] = ghApp.Client
	}

	if err := bindGitHubAppToWorkflow(cfg.Repos, ghs); err != nil {
		return nil, err
	}

	// initialize handler
	return &Handler{
		logger: logger,
		ctrl:   controller.New(cfg, logger, osEnv, ghApps),
	}, nil
}

func readConfig(cfg *config.Config, osEnv osenv.OSEnv) error {
	cfgS := osEnv.Getenv("CONFIG")
	if cfgS == "" {
		return errors.New("environment variable 'CONFIG' is required")
	}
	if err := yaml.Unmarshal([]byte(cfgS), cfg); err != nil {
		return fmt.Errorf("parse the configuration as YAML: %w", err)
	}
	return nil
}

func bindGitHubAppToWorkflow(repos []*config.Repo, ghs map[string]*github.Client) error {
	numRepos := len(repos)
	for i := 0; i < numRepos; i++ {
		repo := repos[i]
		numEvents := len(repo.Events)
		gh, ok := ghs[repo.WorkflowGitHubAppName]
		if !ok {
			return errors.New("invalid github app name")
		}
		repo.GitHub = gh
		for j := 0; j < numEvents; j++ {
			ev := repo.Events[j]
			wfCfg := ev.Workflow
			wfCfg.GitHub = gh
		}
	}
	return nil
}
