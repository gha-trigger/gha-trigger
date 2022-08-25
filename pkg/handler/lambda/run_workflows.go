package lambda

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/domain"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) runWorkflows(ctx context.Context, logger *zap.Logger, body interface{}, repo *github.Repository, workflows []*config.WorkflowConfig) (*Response, error) {
	if len(workflows) == 0 {
		return nil, nil //nolint:nilnil
	}

	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	ref := ""
	sha := ""
	switch ev := body.(type) {
	case *github.PullRequestTargetEvent:
		base := ev.GetPullRequest().GetBase()
		ref = base.GetRef()
		sha = base.GetSHA()
	case *github.PushEvent:
		ref = ev.GetRef()
		sha = ev.GetAfter()
	case *github.ReleaseEvent:
		release := ev.GetRelease()
		ref = fmt.Sprintf("refs/tags/%s", release.GetTagName())
		// TODO sha
	case *github.StatusEvent:
		// ref n/a
		sha = ev.GetSHA()
	default:
		if hasRef, ok := body.(domain.HasRef); ok {
			ref = hasRef.GetRef()
		} else if hasDeployment, ok := body.(domain.HasDeployment); ok {
			deploy := hasDeployment.GetDeployment()
			ref = deploy.GetRef()
			sha = deploy.GetSHA()
		} else if hasPR, ok := body.(domain.HasPR); ok {
			// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
			// sha: Last merge commit on the GITHUB_REF branch
			// ref: PR merge branch refs/pull/:prNumber/merge
			pr, err := handler.waitPRMergeable(ctx, hasPR.GetPullRequest(), repoOwner, repoName)
			if err != nil {
				return nil, err
			}
			if !pr.GetMergeable() {
				return &Response{
					StatusCode: http.StatusBadRequest,
					Body: map[string]interface{}{
						"error": "pull_request isn't mergeable",
					},
				}, nil
			}
			ref = fmt.Sprintf("refs/pull/%v/merge", pr.GetNumber())
			sha = pr.GetMergeCommitSHA()
		}
		// TODO go-github doesn't support registry_package event
	}

	numWorkflows := len(workflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := workflows[i]
		// Run GitHub Actions Workflow
		inputs := make(map[string]interface{}, len(workflow.Inputs))
		for k, v := range workflow.Inputs {
			inputs[k] = v
			inputs["payload"] = body
		}
		inputs["repo_owner"] = repoOwner
		inputs["repo_name"] = repoName
		inputs["sha"] = sha
		inputs["ref"] = ref
		_, err := handler.gh.RunWorkflow(ctx, workflow.RepoOwner, workflow.RepoName, workflow.WorkflowFileName, github.CreateWorkflowDispatchEventRequest{
			Ref:    workflow.Ref,
			Inputs: inputs,
		})
		if err != nil {
			logger.Error(
				"create a workflow dispatch event by file name",
				zap.Error(err),
				zap.String("workflow_repo_owner", workflow.RepoOwner),
				zap.String("workflow_repo_name", workflow.RepoName),
				zap.String("workflow_file_name", workflow.WorkflowFileName))
			continue
		}
	}
	return nil, nil //nolint:nilnil
}

func (handler *Handler) waitPRMergeable(ctx context.Context, pr *github.PullRequest, repoOwner, repoName string) (*github.PullRequest, error) {
	for i := 0; i < 10; i++ {
		if m := pr.Mergeable; m != nil {
			return pr, nil
		}
		// polling
		if err := wait(ctx, 10*time.Second); err != nil { //nolint:gomnd
			return nil, err
		}
		p, _, err := handler.gh.GetPR(ctx, repoOwner, repoName, pr.GetNumber())
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
