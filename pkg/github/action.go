package github

import (
	"context"
)

type ActionsService interface {
	CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event CreateWorkflowDispatchEventRequest) (*Response, error)
}
