package lambda

import (
	"context"
	"strings"

	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
)

func (handler *Handler) handleSlashCommand(ctx context.Context, event *Event, body interface{}) (*Response, error) {
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
	return nil, nil //nolint:nilnil
}
