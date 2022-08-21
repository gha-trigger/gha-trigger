package lambda

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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
	sess := session.Must(session.NewSession())
	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion(cfg.AWS.Region))
	secret, err := readSecretFromSecretsManager(ctx, svc, cfg.AWS.SecretsManager)
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

func (handler *Handler) Do(ctx context.Context, event *Event) (*Response, error) { //nolint:cyclop
	logger := handler.logger
	if err := github.ValidateSignature(event.Headers.Signature, []byte(event.Body), []byte(handler.secret.WebhookSecret)); err != nil {
		logger.Warn("validate the webhook signature", zap.Error(err))
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}, nil
	}

	body, err := github.ParseWebHook(event.Headers.Event, []byte(event.Body))
	if err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}, nil
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

	if issueCommentEvent, ok := body.(*github.IssueCommentEvent); ok {
		cmt := issueCommentEvent.GetComment()
		htmlURL := cmt.GetHTMLURL()
		if !strings.Contains(htmlURL, "/issue/") {
			cmtBody := cmt.GetBody()
			if strings.HasPrefix(cmtBody, "/rerun-workflow") {
				return handler.rerunWorkflows(ctx, event, cmtBody)
			}
			if strings.HasPrefix(cmtBody, "/rerun-failed-job") {
				return handler.rerunFailedJobs(ctx, event, cmtBody)
			}
			if strings.HasPrefix(cmtBody, "/rerun-job") {
				return handler.rerunJobs(ctx, event, cmtBody)
			}
			if strings.HasPrefix(cmtBody, "/cancel") {
				return handler.cancelWorkflows(ctx, event, cmtBody)
			}
		}
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
	if len(workflows) == 0 {
		return nil, nil //nolint:nilnil
	}

	numWorkflows := len(workflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := workflows[i]
		// Run GitHub Actions Workflow
		inputs := make(map[string]interface{}, len(workflow.Inputs))
		for k, v := range workflow.Inputs {
			inputs[k] = v
			inputs["payload"] = body
		}
		_, err := handler.gh.RunWorkflow(ctx, workflow.RepoOwner, workflow.RepoName, workflow.WorkflowFileName, github.CreateWorkflowDispatchEventRequest{
			Ref:    workflow.Ref,
			Inputs: inputs,
		})
		if err != nil {
			logger.Error(
				"create a workflow dispatch event by file name",
				zap.Error(err),
				zap.String("workflow_repo_owner", workflow.RepoOwner),
				zap.String("workflow_repo_name", workflow.RepoName),
				zap.String("workflow_file_name", workflow.WorkflowFileName))
			continue
		}
	}
	return nil, nil //nolint:nilnil
}

func (handler *Handler) setGitHub(param *github.ParamNewApp) error {
	if handler.gh != nil {
		return nil
	}
	return handler.refreshGitHub(param)
}

func (handler *Handler) refreshGitHub(param *github.ParamNewApp) error {
	gh, err := github.NewApp(param)
	if err != nil {
		return err
	}
	handler.gh = gh
	return nil
}
