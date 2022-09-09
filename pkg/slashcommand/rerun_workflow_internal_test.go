package slashcommand

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type workflowRerunner struct {
	resp *github.Response
	err  error
}

func (c *workflowRerunner) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error) {
	return c.resp, c.err
}

func Test_rerunWorkflows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
		matched bool
		owner   string
		repo    string
		body    string
		gh      WorkflowRerunner
	}{
		{
			name: "other command",
			body: "/rerun-workflow-foo",
		},
		{
			name:    "ids are required",
			body:    "/rerun-workflow",
			matched: true,
		},
		{
			name:    "invalid id",
			body:    "/rerun-workflow 1 foo",
			matched: true,
		},
		{
			name:    "normal",
			body:    "/rerun-workflow 1 2",
			matched: true,
			gh:      &workflowRerunner{},
		},
	}
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matched, err := rerunWorkflows(ctx, logger, tt.gh, tt.owner, tt.repo, tt.body)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if matched != tt.matched {
				t.Fatalf("wanted %v, got %v", tt.matched, matched)
			}
		})
	}
}
