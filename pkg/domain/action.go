package domain

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/github"
)

type ActionsService interface {
	CreateWorkflowDispatchEventByFileName(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error)
}

type GitHub interface {
	GetPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error)
	ListPRFiles(ctx context.Context, param *github.ParamsListPRFiles) ([]*github.CommitFile, *github.Response, error)
	CancelWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
	RunWorkflow(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error)
	RerunJob(ctx context.Context, owner, repo string, jobID int64) (*github.Response, error)
	RerunWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
	RerunFailedJobs(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
}

type HasPR interface {
	GetPullRequest() *github.PullRequest
}

type HasRef interface {
	GetRef() string
}

type HasDeployment interface {
	GetDeployment() *github.Deployment
}

type Event struct {
	Body            interface{}
	ChangedFiles    []string
	ChangedFileObjs []*github.CommitFile
	Repo            *github.Repository
	Type            string
	Action          string
	Request         *Request
}

type Request struct {
	// Generate template > Method request passthrough
	Body   string              `json:"body-json"`
	Params *RequestParamsField `json:"params"`
}

type RequestParamsField struct {
	Headers map[string]string `json:"header"`
}

type Response struct {
	StatusCode int              `json:"statusCode"`
	Headers    *ResponseHeaders `json:"headers"`
	Body       interface{}      `json:"body"`
}

type ResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}
