package lambda

import (
	"context"
	"fmt"
	"net/http"

	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/aws"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/config"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/domain"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Handler struct {
	secret *config.Secret
	gh     domain.GitHub
	cfg    *config.Config
	logger *zap.Logger
	osEnv  osenv.OSEnv
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
	secret, err := awsClient.GetSecret(ctx, cfg.AWS.SecretsManager)
	if err != nil {
		return nil, fmt.Errorf("read the secret value from AWS Secrets Manager: %w", err)
	}
	// initialize handler
	return &Handler{
		cfg:    cfg,
		osEnv:  osEnv,
		logger: logger,
		secret: secret,
	}, nil
}

type Event struct {
	Body         string             `json:"body"`
	Headers      *Headers           `json:"headers"`
	ChangedFiles []string           `json:"-"`
	Repo         *github.Repository `json:"-"`
}

type Headers struct {
	Signature string `json:"signature"`
	Event     string `json:"event"`
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

func (handler *Handler) Do(ctx context.Context, event *Event) (*Response, error) {
	logger := handler.logger
	body, resp := handler.validate(logger, event)
	if resp != nil {
		return resp, nil
	}

	paramNewApp := &github.ParamNewApp{
		AppID:   handler.cfg.GitHubApp.AppID,
		KeyFile: handler.secret.GitHubAppPrivateKey,
	}
	if hasInstallation, ok := body.(github.HasInstallation); ok {
		paramNewApp.InstallationID = hasInstallation.GetInstallation().GetID()
		if err := handler.setGitHub(paramNewApp); err != nil {
			logger.Error("set a GitHub Client", zap.Error(err))
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"error": "Internal Server Error",
				},
			}, nil
		}
	}

	return handler.do(ctx, logger, event, body)
}

func (handler *Handler) do(ctx context.Context, logger *zap.Logger, event *Event, body interface{}) (*Response, error) {
	if resp, err := handler.handleSlashCommand(ctx, logger, body); resp != nil {
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
	event.Repo = repo
	repoOwner := repo.GetOwner()
	logger = logger.With(
		zap.String("event_repo_owner", repoOwner.GetLogin()),
		zap.String("event_repo_name", repo.GetName()),
	)

	// route and filter request
	// list labels and changed files
	workflows, resp, err := handler.match(body, event)
	if err != nil {
		return resp, err
	}

	return handler.runWorkflows(ctx, logger, body, workflows)
}
