package slashcommand

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"go.uber.org/zap"
)

func Handle(ctx context.Context, logger *zap.Logger, repoCfg *config.Repo, ev *domain.Event) bool {
	if ev.Type != "issue_comment" {
		return false
	}
	cmt := ev.Payload.Comment
	if strings.Contains(cmt.GetHTMLURL(), "/issue/") {
		return false
	}

	words := strings.Split(cmt.GetBody(), " ")
	firstWord := words[0]
	switch firstWord {
	case "/rerun-workflow":
		rerunWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words[1:])
		return true
	case "/rerun-failed-job":
		rerunFailedJobs(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words[1:])
		return true
	case "/cancel":
		cancelWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words[1:])
		return true
	case "/rerun-job":
		rerunJobs(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words[1:])
		return true
	}
	return false
}
