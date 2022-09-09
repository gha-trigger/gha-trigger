package slashcommand

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type canceler struct {
	resp *github.Response
	err  error
}

func (c *canceler) CancelWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error) {
	return c.resp, c.err
}

func Test_cancelWorkflows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		owner string
		repo  string
		words []string
		gh    WorkflowCanceler
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
			gh:    &canceler{},
		},
	}
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cancelWorkflows(ctx, logger, tt.gh, tt.owner, tt.repo, tt.words)
		})
	}
}
