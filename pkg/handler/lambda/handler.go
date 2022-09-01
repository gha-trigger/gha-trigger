package lambda

import (
	"context"
	"net/http"
	"strings"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type Event struct {
	Body         interface{}
	ChangedFiles []string
	Repo         *github.Repository
	Type         string
	Action       string
	Request      *Request
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

type Request struct {
	// Generate template > Method request passthrough
	Body   string              `json:"body-json"`
	Params *RequestParamsField `json:"params"`
}

type RequestParamsField struct {
	Headers map[string]string `json:"header"`
}

func (handler *Handler) Do(ctx context.Context, req *Request) (*Response, error) {
	// func (handler *Handler) Do(ctx context.Context, e interface{}) (*Response, error) {
	// 	logger := handler.logger
	// 	logger.Info("start a request")
	// 	defer logger.Info("end a request")
	//
	// 	var req *Request
	//
	// 	if err := json.NewEncoder(os.Stderr).Encode(e); err != nil {
	// 		return nil, err
	// 	}
	logger := handler.logger
	logger.Info("start a request")
	defer logger.Info("end a request")

	// Normalize headers
	headers := make(map[string]string, len(req.Params.Headers))
	for k, v := range req.Params.Headers {
		headers[strings.ToUpper(k)] = v
	}
	req.Params.Headers = headers

	ghApp, ev, resp := handler.validate(logger, req)

	if resp != nil {
		return resp, nil
	}

	return handler.do(ctx, logger, ghApp, ev)
}

type hasAction interface {
	GetAction() string
}

func (handler *Handler) do(ctx context.Context, logger *zap.Logger, ghApp *GitHubApp, ev *Event) (*Response, error) {
	body := ev.Body

	repoEvent, ok := body.(RepoEvent)
	if !ok {
		logger.Info("event is ignored because a repository isn't found in the payload")
		return &Response{
			StatusCode: http.StatusOK,
			Body: map[string]interface{}{
				"message": "event is ignored because a repository isn't found in the payload",
			},
		}, nil
	}

	ghRepo := repoEvent.GetRepo()
	ev.Repo = ghRepo
	var repoCfg *config.Repo
	for _, repo := range handler.cfg.Repos {
		if repo.RepoOwner != ghRepo.GetOwner().GetLogin() || repo.RepoName != ghRepo.GetName() {
			continue
		}
		repoCfg = repo
		break
	}
	if repoCfg == nil {
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"message": "repository config isn't found",
			},
		}, nil
	}

	logger = logger.With(
		zap.String("event_repo_owner", repoCfg.RepoOwner),
		zap.String("event_repo_name", repoCfg.RepoName),
	)

	if resp, err := handler.handleSlashCommand(ctx, logger, repoCfg, body); resp != nil {
		return resp, err
	}

	if actionEvent, ok := body.(hasAction); ok {
		ev.Action = actionEvent.GetAction()
	}

	// route and filter request
	// list labels and changed files
	workflows, resp, err := handler.match(ev, repoCfg)
	if err != nil {
		return resp, err
	}

	return handler.runWorkflows(ctx, logger, ghApp.Client, ev, repoCfg, workflows)
}
