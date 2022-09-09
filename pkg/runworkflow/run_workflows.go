package runworkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	PullRequest  *github.PullRequest  `json:"pull_request,omitempty"`
}

func getWorkflowInput(ev *domain.Event) (map[string]interface{}, error) {
	input := &WorkflowInput{
		Event:        ev.Raw,
		EventName:    ev.Type,
		ChangedFiles: ev.ChangedFileObjs,
		PullRequest:  ev.Payload.PullRequest,
	}

	b, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal input as JSON: %w", err)
	}
	return map[string]interface{}{
		"data": string(b),
	}, nil
}

func RunWorkflows(ctx context.Context, logger *zap.Logger, gh *github.Client, ev *domain.Event, repoCfg *config.Repo, workflows []*config.Workflow) error {
	if len(workflows) == 0 {
		logger.Info("no workflow is run")
		return nil //nolint:nilnil
	}

	repo := ev.Payload.Repo
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	if pr := ev.Payload.PullRequest; pr != nil {
		// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
		// sha: Last merge commit on the GITHUB_REF branch
		// ref: PR merge branch refs/pull/:prNumber/merge
		pr, err := waitPRMergeable(ctx, gh, pr, repoOwner, repoName)
		if err != nil {
			return fmt.Errorf("wait until pull request's mergeable becomes not nil: %w", err)
		}
		if !pr.GetMergeable() {
			logger.Warn("pull_request isn't mergeable")
			return nil
		}
		ev.Payload.PullRequest = pr
	}

	inputs, err := getWorkflowInput(ev)
	if err != nil {
		return err
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
	return nil
}

func waitPRMergeable(ctx context.Context, gh *github.Client, pr *github.PullRequest, repoOwner, repoName string) (*github.PullRequest, error) {
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
