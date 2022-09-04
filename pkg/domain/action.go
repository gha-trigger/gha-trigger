package domain

import (
	"context"
	"errors"
	"fmt"

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
	Raw             map[string]interface{}
	ChangedFiles    []string
	ChangedFileObjs []*github.CommitFile
	Repo            *github.Repository
	Type            string
	Action          string
	Request         *Request
	GitHub          *github.Client
}

func getChangedFiles(files []*github.CommitFile) []string {
	changedFileMap := make(map[string]struct{}, len(files))
	for _, file := range files {
		if f := file.GetFilename(); f != "" {
			changedFileMap[f] = struct{}{}
		}
		if f := file.GetPreviousFilename(); f != "" {
			changedFileMap[f] = struct{}{}
		}
	}
	changedFiles := make([]string, 0, len(changedFileMap))
	for k := range changedFileMap {
		changedFiles = append(changedFiles, k)
	}
	return changedFiles
}

func (ev *Event) GetChangedFiles(ctx context.Context) ([]string, error) {
	if ev.ChangedFileObjs == nil {
		switch ev.Type {
		case "pull_request", "pull_request_target":
			hasPR, ok := ev.Body.(HasPR)
			if !ok {
				return nil, errors.New("body must be HasPR")
			}
			pr := hasPR.GetPullRequest()
			files, _, err := ev.GitHub.ListPRFiles(ctx, &github.ParamsListPRFiles{
				Owner:  ev.Repo.GetOwner().GetLogin(),
				Repo:   ev.Repo.GetName(),
				Number: pr.GetNumber(),
				Count:  pr.GetChangedFiles(),
			})
			if err != nil {
				return nil, fmt.Errorf("list pull request files: %w", err)
			}
			ev.ChangedFileObjs = files
			ev.ChangedFiles = getChangedFiles(files)
		default:
		}
	}
	return ev.ChangedFiles, nil
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
