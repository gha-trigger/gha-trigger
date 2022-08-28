package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
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

	numEvents := len(cfg.Events)
	for i := 0; i < numEvents; i++ {
		evCfg := cfg.Events[i]
		numWorkflows := len(evCfg.Workflows)
		for j := 0; j < numWorkflows; j++ {
			wfCfg := evCfg.Workflows[j]
			gh, ok := ghs[wfCfg.GitHubAppName]
			if !ok {
				return nil, errors.New("invalid github app name")
			}
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

type Event struct {
	Body         string                         `json:"body"`
	ChangedFiles []string                       `json:"-"`
	Repo         *github.Repository             `json:"-"`
	Event        *events.APIGatewayProxyRequest `json:"-"`
}

type Response struct {
	StatusCode int              `json:"statusCode"`
	Headers    *ResponseHeaders `json:"headers"`
	Body       interface{}      `json:"body"`
}

type ResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type RepoEvent interface {
	GetRepo() *github.Repository
}

type HasEventType interface {
	GetAction() string
}

func (handler *Handler) Do(ctx context.Context, event *events.APIGatewayProxyRequest) (*Response, error) {
	logger := handler.logger
	ghApp, body, resp := handler.validate(logger, event)

	if resp != nil {
		return resp, nil
	}

	return handler.do(ctx, logger, ghApp, body)
}

func (handler *Handler) do(ctx context.Context, logger *zap.Logger, ghApp *GitHubApp, body interface{}) (*Response, error) {
	if resp, err := handler.handleSlashCommand(ctx, logger, ghApp.Client, body); resp != nil {
		return resp, err
	}

	repoEvent, ok := body.(RepoEvent)
	if !ok {
		return &Response{
			StatusCode: http.StatusOK,
			Body: map[string]interface{}{
				"message": "event is ignored because a repository isn't found in the payload",
			},
		}, nil
	}
	repo := repoEvent.GetRepo()
	ev := &Event{
		Repo: repo,
	}
	repoOwner := repo.GetOwner()
	logger = logger.With(
		zap.String("event_repo_owner", repoOwner.GetLogin()),
		zap.String("event_repo_name", repo.GetName()),
	)

	// route and filter request
	// list labels and changed files
	workflows, resp, err := handler.match(body, ev)
	if err != nil {
		return resp, err
	}

	return handler.runWorkflows(ctx, logger, ghApp.Client, body, repo, workflows)
}
