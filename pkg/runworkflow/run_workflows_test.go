package runworkflow_test

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/runworkflow"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"go.uber.org/zap"
)

type githubPRClient struct {
	pr   *github.PullRequest
	resp *github.Response
	err  error
}

func (client *githubPRClient) GetPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return client.pr, client.resp, client.err
}

type githubWorkflowClient struct {
	resp *github.Response
	err  error
}

func (client *githubWorkflowClient) RunWorkflow(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error) {
	return client.resp, client.err
}

func TestRunWorkflows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		wantErr   bool
		ev        *domain.Event
		repoCfg   *config.Repo
		workflows []*config.Workflow
		gh        runworkflow.GitHubPRClient
	}{
		{
			name: "pull_request",
			ev: &domain.Event{
				Type: "pull_request",
				Raw:  map[string]interface{}{},
				Payload: &domain.Payload{
					Repo: &github.Repository{
						Name: util.StrP("example-main"),
						Owner: &github.User{
							Login: util.StrP("gha-trigger"),
						},
					},
					PullRequest: &github.PullRequest{},
				},
			},
			repoCfg: &config.Repo{
				RepoOwner:  "gha-trigger",
				CIRepoName: "example-ci",
			},
			workflows: []*config.Workflow{
				{
					WorkflowFileName: "test_pull_request.yaml",
					Ref:              "pull_request",
					GitHub:           &githubWorkflowClient{},
				},
			},
			gh: &githubPRClient{
				pr: &github.PullRequest{
					Mergeable: util.BoolP(true),
				},
			},
		},
	}
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := runworkflow.RunWorkflows(ctx, logger, tt.gh, tt.ev, tt.repoCfg, tt.workflows); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
