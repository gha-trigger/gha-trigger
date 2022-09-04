package lambda

import (
	"context"
	"net/http"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/route"
	"go.uber.org/zap"
)

func (handler *Handler) Do(ctx context.Context, req *domain.Request) (*domain.Response, error) {
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

func (handler *Handler) do(ctx context.Context, logger *zap.Logger, ghApp *GitHubApp, ev *domain.Event) (*domain.Response, error) {
	body := ev.Body

	if ev.Payload.Repo == nil {
		logger.Info("event is ignored because a repository isn't found in the payload")
		return &domain.Response{
			StatusCode: http.StatusOK,
			Body: map[string]interface{}{
				"message": "event is ignored because a repository isn't found in the payload",
			},
		}, nil
	}

	ghRepo := ev.Payload.Repo
	var repoCfg *config.Repo
	for _, repo := range handler.cfg.Repos {
		if repo.RepoOwner != ghRepo.GetOwner().GetLogin() || repo.RepoName != ghRepo.GetName() {
			continue
		}
		repoCfg = repo
		break
	}
	if repoCfg == nil {
		return &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"message": "repository config isn't found",
			},
		}, nil
	}

	logger = logger.With(
		zap.String("event_repo_owner", repoCfg.RepoOwner),
		zap.String("event_repo_name", repoCfg.RepoName),
		zap.String("ci_repo_name", repoCfg.CIRepoName),
	)

	if resp, err := handler.handleSlashCommand(ctx, logger, repoCfg, body); resp != nil {
		return resp, err
	}

	// route and filter request
	// list labels and changed files
	workflows, resp, err := route.Match(ctx, ev, repoCfg)
	if err != nil {
		return resp, err
	}

	return handler.runWorkflows(ctx, logger, ghApp.Client, ev, repoCfg, workflows)
}
