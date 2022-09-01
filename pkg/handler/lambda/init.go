package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/aws"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	cfg    *config.Config
	logger *zap.Logger
	osEnv  osenv.OSEnv
	ghs    map[int64]*GitHubApp
}

type GitHubApp struct {
	Name          string
	WebhookSecret string
	Client        *github.Client
}

func newGitHubApp(ctx context.Context, awsClient *aws.Client, appCfg *config.GitHubApp) (*GitHubApp, error) {
	paramNewApp := &github.ParamNewApp{
		AppID:          appCfg.AppID,
		InstallationID: appCfg.InstallationID,
		Org:            appCfg.Org,
		User:           appCfg.User,
	}
	input := &aws.GetSecretValueInput{
		SecretId: aws.String(appCfg.Secret.SecretID),
	}
	secretOutput, err := awsClient.GetSecretValueWithContext(ctx, input) //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("read the secret value from AWS Secrets Manager: %w", err)
	}
	secret := &config.GitHubAppSecret{}
	if err := json.Unmarshal([]byte(*secretOutput.SecretString), secret); err != nil {
		return nil, fmt.Errorf("unmarshal the GitHub App Secret as JSON: %w", err)
	}
	paramNewApp.KeyFile = secret.GitHubAppPrivateKey
	if secret.AppID != 0 {
		paramNewApp.AppID = secret.AppID
	}
	if secret.InstallationID != 0 {
		paramNewApp.InstallationID = secret.InstallationID
	}
	gh, err := github.NewApp(ctx, paramNewApp)
	if err != nil {
		return nil, fmt.Errorf("create a GitHub Client: %w", err)
	}
	return &GitHubApp{
		Name:          appCfg.Name,
		WebhookSecret: secret.WebhookSecret,
		Client:        gh,
	}, nil
}

func New(ctx context.Context, logger *zap.Logger) (*Handler, error) {
	// read config
	cfg := &config.Config{}
	osEnv := osenv.New()
	if err := yaml.Unmarshal([]byte(osEnv.Getenv("CONFIG")), cfg); err != nil {
		return nil, fmt.Errorf("parse the configuration as YAML: %w", err)
	}
	initCfg(cfg)
	// read env
	// read secret
	awsClient := aws.New(cfg.AWS)
	numGitHubApps := len(cfg.GitHubApps)
	ghApps := make(map[int64]*GitHubApp, numGitHubApps)
	ghs := make(map[string]*github.Client, numGitHubApps)
	for i := 0; i < numGitHubApps; i++ {
		appCfg := cfg.GitHubApps[i]
		ghApp, err := newGitHubApp(ctx, awsClient, appCfg)
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

func initCfg(cfg *config.Config) {
	compileCfg(cfg)
	for _, repo := range cfg.Repos {
		for _, event := range repo.Events {
			for _, match := range event.Matches {
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
}

func compileCfg(cfg *config.Config) {
	for _, repo := range cfg.Repos {
		for _, event := range repo.Events {
			for _, match := range event.Matches {
				match.CompiledBranches = compileStrings(match.Branches)
				match.CompiledTags = compileStrings(match.Tags)
				match.CompiledPaths = compileStrings(match.Paths)
				match.CompiledBranchesIgnore = compileStrings(match.BranchesIgnore)
				match.CompiledTagsIgnore = compileStrings(match.TagsIgnore)
				match.CompiledPathsIgnore = compileStrings(match.PathsIgnore)
			}
		}
	}
}

func compileStrings(list []string) []*regexp.Regexp {
	n := len(list)
	if n == 0 {
		return nil
	}
	arr := make([]*regexp.Regexp, 0, n)
	for _, branch := range list {
		c, err := regexp.Compile(branch)
		if err == nil {
			arr = append(arr, c)
		}
	}
	return arr
}
