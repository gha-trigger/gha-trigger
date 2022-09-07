package route

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func Test_matchTags(t *testing.T) {
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
			name:        "no tags",
			matchConfig: &config.Match{},
			exp:         true,
		},
		{
			name: "ref not match",
			matchConfig: &config.Match{
				Tags: []*config.StringMatch{
					{
						Type:  "equal",
						Value: "main",
					},
				},
			},
			event: &domain.Event{
				Payload: &domain.Payload{
					Ref: "refs/tags/develop",
				},
			},
		},
		{
			name: "ref match",
			exp:  true,
			matchConfig: &config.Match{
				Tags: []*config.StringMatch{
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
					Ref: "refs/tags/main",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, _, err := matchTags(ctx, tt.matchConfig, tt.event)
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

func Test_matchTagsIgnore(t *testing.T) {
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
			name: "ref not match",
			matchConfig: &config.Match{
				TagsIgnore: []*config.StringMatch{
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
					Ref: "refs/tags/main",
				},
			},
		},
		{
			name: "ref match",
			exp:  true,
			matchConfig: &config.Match{
				TagsIgnore: []*config.StringMatch{
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
					Ref: "refs/tags/main",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, _, err := matchTagsIgnore(ctx, tt.matchConfig, tt.event)
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
