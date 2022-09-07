package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/aws"
	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	cfg    *config.Config
	logger *zap.Logger
	osEnv  osenv.OSEnv
	ghs    map[int64]*githubapp.GitHubApp
}

func New(ctx context.Context, logger *zap.Logger) (*Handler, error) {
	// read config
	cfg := &config.Config{}
	osEnv := osenv.New()
	cfgS := osEnv.Getenv("CONFIG")
	if cfgS == "" {
		return nil, errors.New("environment variable 'CONFIG' is required")
	}
	if err := yaml.Unmarshal([]byte(cfgS), cfg); err != nil {
		return nil, fmt.Errorf("parse the configuration as YAML: %w", err)
	}
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration is invalid: %w", err)
	}
	if err := initCfg(cfg); err != nil {
		return nil, fmt.Errorf("initialize configuration: %w", err)
	}
	// read env
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

	numRepos := len(cfg.Repos)
	for i := 0; i < numRepos; i++ {
		repo := cfg.Repos[i]
		numEvents := len(repo.Events)
		gh, ok := ghs[repo.WorkflowGitHubAppName]
		if !ok {
			return nil, errors.New("invalid github app name")
		}
		repo.GitHub = gh
		for j := 0; j < numEvents; j++ {
			ev := repo.Events[j]
			wfCfg := ev.Workflow
			wfCfg.GitHub = gh
		}
	}

	// initialize handler
	return &Handler{
		cfg:    cfg,
		osEnv:  osEnv,
		logger: logger,
		ghs:    ghApps,
	}, nil
}

func initCfg(cfg *config.Config) error {
	for _, repo := range cfg.Repos {
		for _, event := range repo.Events {
			for _, match := range event.Matches {
				if err := match.Compile(); err != nil {
					return err
				}
				for _, ev := range match.Events {
					if ev.Name == "pull_request" && ev.Types == nil {
						// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
						// > By default, a workflow only runs when a pull_request event's activity type is
						// > opened, synchronize, or reopened.
						ev.Types = []string{"opened", "synchronize", "reopened"}
					}
				}
			}
		}
	}
	return nil
}
