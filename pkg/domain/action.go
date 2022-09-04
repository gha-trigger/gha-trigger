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

type Event struct {
	Body            interface{}
	Raw             map[string]interface{}
	Payload         *Payload
	ChangedFiles    []string
	ChangedFileObjs []*github.CommitFile
	Type            string
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
			if ev.Payload.PullRequest == nil {
				return nil, errors.New("body must have a pull request")
			}
			pr := ev.Payload.PullRequest
			files, _, err := ev.GitHub.ListPRFiles(ctx, &github.ParamsListPRFiles{
				Owner:  ev.Payload.Repo.GetOwner().GetLogin(),
				Repo:   ev.Payload.Repo.GetName(),
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

type Payload struct {
	Repo        *github.Repository  `json:"repository"`
	PullRequest *github.PullRequest `json:"pull_request"`
	Ref         string              `json:"ref"`
	Action      string              `json:"action"`
}
