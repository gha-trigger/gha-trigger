package github

import (
	"context"
)

type ActionsService interface {
	CancelWorkflowRunByID(ctx context.Context, owner, repo string, runID int64) (*Response, error)
	CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event CreateWorkflowDispatchEventRequest) (*Response, error)
	RerunJobByID(ctx context.Context, owner, repo string, jobID int64) (*Response, error)
	RerunFailedJobsByID(ctx context.Context, owner, repo string, runID int64) (*Response, error)
	RerunWorkflowByID(ctx context.Context, owner, repo string, runID int64) (*Response, error)
}

func (client *Client) RunWorkflow(ctx context.Context, owner, repo, workflowFileName string, event CreateWorkflowDispatchEventRequest) (*Response, error) {
	return client.action.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowFileName, event)
}

func (client *Client) RerunJob(ctx context.Context, owner, repo string, jobID int64) (*Response, error) {
	return client.action.RerunJobByID(ctx, owner, repo, jobID)
}

func (client *Client) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return client.action.RerunWorkflowByID(ctx, owner, repo, runID)
}

func (client *Client) RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return client.action.RerunFailedJobsByID(ctx, owner, repo, runID)
}

func (client *Client) CancelWorkflow(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return client.action.CancelWorkflowRunByID(ctx, owner, repo, runID)
}
