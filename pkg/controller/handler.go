package controller

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/gha-trigger/gha-trigger/pkg/route"
	"github.com/gha-trigger/gha-trigger/pkg/runworkflow"
	"github.com/gha-trigger/gha-trigger/pkg/slashcommand"
	"go.uber.org/zap"
)

func (ctrl *Controller) Do(ctx context.Context, logger *zap.Logger, req *domain.Request) error {
	toUpperHeaders(req.Params)

	ghApp, ev, err := ctrl.validate(logger, req)
	if err != nil {
		return err
	}
	logger = logger.With(zap.String("event_type", ev.Type))

	return ctrl.do(ctx, logger, ghApp, ev)
}

func toUpperHeaders(params *domain.RequestParamsField) {
	// Normalize headers
	headers := make(map[string]string, len(params.Headers))
	for k, v := range params.Headers {
		headers[strings.ToUpper(k)] = v
	}
	params.Headers = headers
}

func getRepoConfig(ghRepo *github.Repository, repos []*config.Repo) *config.Repo {
	for _, repo := range repos {
		if repo.RepoOwner == ghRepo.GetOwner().GetLogin() && repo.RepoName == ghRepo.GetName() {
			return repo
		}
	}
	return nil
}

func (ctrl *Controller) do(ctx context.Context, logger *zap.Logger, ghApp *githubapp.GitHubApp, ev *domain.Event) error {
	if ev.Payload.Repo == nil {
		logger.Info("event is ignored because a repository isn't found in the payload")
		return nil
	}

	repoCfg := getRepoConfig(ev.Payload.Repo, ctrl.cfg.Repos)
	if repoCfg == nil {
		logger.Error("repository config isn't found")
		return nil
	}

	logger = logger.With(
		zap.String("event_repo_owner", repoCfg.RepoOwner),
		zap.String("event_repo_name", repoCfg.RepoName),
		zap.String("ci_repo_name", repoCfg.CIRepoName),
	)

	if slashcommand.Handle(ctx, logger, repoCfg, ev) {
		return nil
	}

	// route and filter request
	// list labels and changed files
	workflows, err := route.Match(ctx, ev, repoCfg)
	if err != nil {
		return err
	}

	return runworkflow.RunWorkflows(ctx, logger, ghApp.Client, ev, repoCfg, workflows)
}
