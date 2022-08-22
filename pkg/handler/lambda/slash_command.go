package lambda

import (
	"context"
	"strings"

	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) handleSlashCommand(ctx context.Context, logger *zap.Logger, body interface{}) (*Response, error) {
	issueCommentEvent, ok := body.(*github.IssueCommentEvent)
	if !ok {
		return nil, nil //nolint:nilnil
	}
	cmt := issueCommentEvent.GetComment()
	if strings.Contains(cmt.GetHTMLURL(), "/issue/") {
		return nil, nil //nolint:nilnil
	}
	cmtBody := cmt.GetBody()
	repo := issueCommentEvent.GetRepo()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	if strings.HasPrefix(cmtBody, "/rerun-workflow") {
		return handler.rerunWorkflows(ctx, logger, owner, repoName, cmtBody)
	}
	if strings.HasPrefix(cmtBody, "/rerun-failed-job") {
		return handler.rerunFailedJobs(ctx, logger, owner, repoName, cmtBody)
	}
	if strings.HasPrefix(cmtBody, "/rerun-job") {
		return handler.rerunJobs(ctx, logger, owner, repoName, cmtBody)
	}
	if strings.HasPrefix(cmtBody, "/cancel") {
		return handler.cancelWorkflows(ctx, logger, owner, repoName, cmtBody)
	}
	return nil, nil //nolint:nilnil
}
