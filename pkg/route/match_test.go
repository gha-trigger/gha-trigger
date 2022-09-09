package route_test

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/route"
	"github.com/google/go-cmp/cmp"
)

func TestMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name    string
		wantErr bool
		exp     []*config.Workflow
		event   *domain.Event
		repo    *config.Repo
	}{
		{
			name: "normal",
			exp: []*config.Workflow{
				{
					WorkflowFileName: "test.yaml",
				},
				{
					WorkflowFileName: "deploy.yaml",
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/heads/main",
				},
			},
			repo: &config.Repo{
				Events: []*config.Event{
					{
						// no maches
						Workflow: &config.Workflow{
							WorkflowFileName: "test.yaml",
						},
					},
					{
						Matches: []*config.Match{
							{
								Branches: []*config.StringMatch{
									{
										Type:  "equal",
										Value: "main",
									},
								},
							},
						},
						Workflow: &config.Workflow{
							WorkflowFileName: "deploy.yaml",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			workflows, err := route.Match(ctx, tt.event, tt.repo)
			if err != nil {
				if tt.wantErr {
					return
				}
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(workflows, tt.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
