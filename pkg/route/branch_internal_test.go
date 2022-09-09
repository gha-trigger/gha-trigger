package route

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/util"
)

func Test_matchBranches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name        string
		wantErr     bool
		exp         bool
		matchConfig *config.Match
		event       *domain.Event
	}{
		{
			name:        "no branches",
			matchConfig: &config.Match{},
			exp:         true,
		},
		{
			name: "pr not match",
			matchConfig: &config.Match{
				Branches: []*config.StringMatch{
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: util.StrP("develop"),
						},
					},
				},
			},
		},
		{
			name: "pr match",
			exp:  true,
			matchConfig: &config.Match{
				Branches: []*config.StringMatch{
					// OR condition
					{
						Type:  "equal",
						Value: "develop",
					},
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: util.StrP("main"),
						},
					},
				},
			},
		},
		{
			name: "ref not match",
			matchConfig: &config.Match{
				Branches: []*config.StringMatch{
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/heads/develop",
				},
			},
		},
		{
			name: "ref match",
			exp:  true,
			matchConfig: &config.Match{
				Branches: []*config.StringMatch{
					// OR condition
					{
						Type:  "equal",
						Value: "develop",
					},
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/heads/main",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, err := matchBranches(ctx, tt.matchConfig, tt.event)
			if err != nil {
				if tt.wantErr {
					return
				}
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if f != tt.exp {
				t.Fatalf("wanted %v, got %v", tt.exp, f)
			}
		})
	}
}

func Test_matchBranchesIgnore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name        string
		wantErr     bool
		exp         bool
		matchConfig *config.Match
		event       *domain.Event
	}{
		{
			name:        "no branches-ignore",
			matchConfig: &config.Match{},
			exp:         true,
		},
		{
			name: "pr not match",
			matchConfig: &config.Match{
				BranchesIgnore: []*config.StringMatch{
					{
						Type:  "equal",
						Value: "develop",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: util.StrP("develop"),
						},
					},
				},
			},
		},
		{
			name: "pr match",
			exp:  true,
			matchConfig: &config.Match{
				BranchesIgnore: []*config.StringMatch{
					// OR condition
					{
						Type:  "equal",
						Value: "develop",
					},
					{
						Type:  "equal",
						Value: "prod",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					PullRequest: &github.PullRequest{
						Base: &github.PullRequestBranch{
							Ref: util.StrP("main"),
						},
					},
				},
			},
		},
		{
			name: "ref not match",
			matchConfig: &config.Match{
				BranchesIgnore: []*config.StringMatch{
					// OR condition
					{
						Type:  "equal",
						Value: "develop",
					},
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/heads/main",
				},
			},
		},
		{
			name: "ref match",
			exp:  true,
			matchConfig: &config.Match{
				BranchesIgnore: []*config.StringMatch{
					// OR condition
					{
						Type:  "equal",
						Value: "develop",
					},
					{
						Type:  "equal",
						Value: "prod",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/heads/main",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, err := matchBranchesIgnore(ctx, tt.matchConfig, tt.event)
			if err != nil {
				if tt.wantErr {
					return
				}
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if f != tt.exp {
				t.Fatalf("wanted %v, got %v", tt.exp, f)
			}
		})
	}
}
