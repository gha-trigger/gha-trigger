package slashcommand

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func Handle(ctx context.Context, logger *zap.Logger, repoCfg *config.Repo, body interface{}) bool {
	issueCommentEvent, ok := body.(*github.IssueCommentEvent)
	if !ok {
		return false
	}
	cmt := issueCommentEvent.GetComment()
	if strings.Contains(cmt.GetHTMLURL(), "/issue/") {
		return false
	}

	words := strings.Split(cmt.GetBody(), " ")
	firstWord := words[0]
	switch firstWord {
	case "/rerun-workflow":
		rerunWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words)
		return true
	case "/rerun-failed-job":
		rerunFailedJobs(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words)
		return true
	case "/cancel":
		cancelWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, words)
		return true
	}
	return false
}
