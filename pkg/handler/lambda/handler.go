package lambda

import (
	"context"
	"net/http"

	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/config"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/domain"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
	"go.uber.org/zap"
)

type Handler struct {
	secret  *Secret
	actions domain.ActionsService
	// gh      domain.GitHub
	cfg    *config.Config
	logger *zap.Logger
}

func New() (*Handler, error) {
	// read config
	// read env
	// read secret
	// initialize handler
	return &Handler{}, nil
}

type Secret struct {
	WebhookSecret string
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
		_, err := handler.actions.CreateWorkflowDispatchEventByFileName(ctx, workflow.RepoOwner, workflow.RepoName, workflow.WorkflowFileName, github.CreateWorkflowDispatchEventRequest{})
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
