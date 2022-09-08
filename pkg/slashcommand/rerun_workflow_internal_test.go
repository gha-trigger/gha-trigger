package slashcommand

import (
	"context"
	"net/http"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/google/go-cmp/cmp"
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
		resp    *domain.Response
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
			name: "ids are required",
			body: "/rerun-workflow",
			resp: &domain.Response{
				StatusCode: http.StatusBadRequest,
				Body: map[string]interface{}{
					"error": "workflow ids are required",
				},
			},
		},
		{
			name: "invalid id",
			body: "/rerun-workflow 1 foo",
			resp: &domain.Response{
				StatusCode: http.StatusBadRequest,
				Body: map[string]interface{}{
					"message": "workflow run id is invalid",
				},
			},
		},
		{
			name: "normal",
			body: "/rerun-workflow 1 2",
			gh:   &workflowRerunner{},
		},
	}
	ctx := context.Background()
	logger, _ := zap.NewProduction()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp, err := rerunWorkflows(ctx, logger, tt.gh, tt.owner, tt.repo, tt.body)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(resp, tt.resp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
