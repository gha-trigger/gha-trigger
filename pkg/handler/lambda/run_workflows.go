package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type WorkflowInput struct {
	// https://docs.github.com/en/actions/learn-github-actions/contexts
	Event        interface{}          `json:"event"`
	EventName    string               `json:"event_name"`
	ChangedFiles []*github.CommitFile `json:"changed_files,omitempty"`
}

func (handler *Handler) getWorkflowInput(logger *zap.Logger, ev *Event) (map[string]interface{}, *Response) {
	body := ev.Body
	input := &WorkflowInput{
		Event:        body,
		EventName:    ev.Type,
		ChangedFiles: ev.ChangedFileObjs,
	}

	b, err := json.Marshal(input)
	if err != nil {
		logger.Error("marshal input as JSON", zap.Error(err))
		return nil, &Response{
			StatusCode: http.StatusInternalServerError,
			Body: map[string]interface{}{
				"error": "Internal Server Error",
			},
		}
	}
	return map[string]interface{}{
		"data": string(b),
	}, nil
}

func (handler *Handler) runWorkflows(ctx context.Context, logger *zap.Logger, gh *github.Client, ev *Event, repoCfg *config.Repo, workflows []*config.Workflow) (*Response, error) {
	if len(workflows) == 0 {
		logger.Info("no workflow is run")
		return nil, nil //nolint:nilnil
	}

	repo := ev.Repo
	body := ev.Body
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	if hasPR, ok := body.(domain.HasPR); ok {
		// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
		// sha: Last merge commit on the GITHUB_REF branch
		// ref: PR merge branch refs/pull/:prNumber/merge
		pr, err := handler.waitPRMergeable(ctx, gh, hasPR.GetPullRequest(), repoOwner, repoName)
		if err != nil {
			logger.Error(
				"wait until pull request's mergeable becomes not nil",
				zap.Error(err))
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"error": "Internal Server Error",
				},
			}, nil
		}
		if !pr.GetMergeable() {
			return &Response{
				StatusCode: http.StatusBadRequest,
				Body: map[string]interface{}{
					"error": "pull_request isn't mergeable",
				},
			}, nil
		}
	}

	inputs, resp := handler.getWorkflowInput(logger, ev)
	if resp != nil {
		return resp, nil
	}

	numWorkflows := len(workflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := workflows[i]
		// Run GitHub Actions Workflow
		logger := logger.With(
			zap.String("workflow_repo_owner", repoCfg.RepoOwner),
			zap.String("workflow_repo_name", repoCfg.CIRepoName),
			zap.String("workflow_file_name", workflow.WorkflowFileName),
			zap.String("workflow_ref", workflow.Ref))
		logger.Info("running a GitHub Actions Workflow")
		_, err := workflow.GitHub.RunWorkflow(ctx, repoCfg.RepoOwner, repoCfg.CIRepoName, workflow.WorkflowFileName, github.CreateWorkflowDispatchEventRequest{
			Ref:    workflow.Ref,
			Inputs: inputs,
		})
		if err != nil {
			logger.Error(
				"create a workflow dispatch event by file name",
				zap.Error(err))
		}
	}
	return nil, nil //nolint:nilnil
}

func (handler *Handler) waitPRMergeable(ctx context.Context, gh *github.Client, pr *github.PullRequest, repoOwner, repoName string) (*github.PullRequest, error) {
	for i := 0; i < 10; i++ {
		if m := pr.Mergeable; m != nil {
			return pr, nil
		}
		// polling
		if err := wait(ctx, 10*time.Second); err != nil { //nolint:gomnd
			return nil, err
		}
		p, _, err := gh.GetPR(ctx, repoOwner, repoName, pr.GetNumber())
		if err != nil {
			return nil, err
		}
		pr = p
	}
	return nil, errors.New("timeout error")
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	}
}
