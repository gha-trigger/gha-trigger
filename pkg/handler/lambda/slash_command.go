package lambda

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) handleSlashCommand(ctx context.Context, logger *zap.Logger, repoCfg *config.Repo, body interface{}) (*domain.Response, error) {
	issueCommentEvent, ok := body.(*github.IssueCommentEvent)
	if !ok {
		return nil, nil //nolint:nilnil
	}
	cmt := issueCommentEvent.GetComment()
	if strings.Contains(cmt.GetHTMLURL(), "/issue/") {
		return nil, nil //nolint:nilnil
	}

	cmtBody := cmt.GetBody()
	if strings.HasPrefix(cmtBody, "/rerun-workflow") {
		return handler.rerunWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, cmtBody)
	}
	if strings.HasPrefix(cmtBody, "/rerun-failed-job") {
		return handler.rerunFailedJobs(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, cmtBody)
	}
	// if strings.HasPrefix(cmtBody, "/rerun-job") {
	// 	return handler.rerunJobs(ctx, logger, gh, owner, repoName, cmtBody)
	// }
	if strings.HasPrefix(cmtBody, "/cancel") {
		return handler.cancelWorkflows(ctx, logger, repoCfg.GitHub, repoCfg.RepoOwner, repoCfg.CIRepoName, cmtBody)
	}
	return nil, nil //nolint:nilnil
}
