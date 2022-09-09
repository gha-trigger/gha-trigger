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
		name  string
		owner string
		repo  string
		words []string
		gh    WorkflowRerunner
	}{
		{
			name: "ids are required",
		},
		{
			name:  "invalid id",
			words: []string{"1", "foo"},
		},
		{
			name:  "normal",
			words: []string{"1", "2"},
			gh:    &workflowRerunner{},
		},
	}
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rerunWorkflows(ctx, logger, tt.gh, tt.owner, tt.repo, tt.words)
		})
	}
}
