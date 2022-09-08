package github

import "context"

type ActionsServiceMock struct {
	Resp *Response
	Err  error
}

func (mock *ActionsServiceMock) CancelWorkflowRunByID(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return mock.Resp, mock.Err
}

func (mock *ActionsServiceMock) CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event CreateWorkflowDispatchEventRequest) (*Response, error) {
	return mock.Resp, mock.Err
}

func (mock *ActionsServiceMock) RerunJobByID(ctx context.Context, owner, repo string, jobID int64) (*Response, error) {
	return mock.Resp, mock.Err
}

func (mock *ActionsServiceMock) RerunFailedJobsByID(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return mock.Resp, mock.Err
}

func (mock *ActionsServiceMock) RerunWorkflowByID(ctx context.Context, owner, repo string, runID int64) (*Response, error) {
	return mock.Resp, mock.Err
}
