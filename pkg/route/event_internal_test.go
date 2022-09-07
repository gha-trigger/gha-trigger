package route

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func Test_matchEventType(t *testing.T) {
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
			name:        "no events",
			matchConfig: &config.Match{},
			exp:         true,
		},
		{
			name: "not match",
			matchConfig: &config.Match{
				Events: []*config.EventType{
					{
						Name: "push",
					},
				},
			},
			event: &domain.Event{
				Type:    "pull_request",
				Payload: &domain.Payload{},
			},
		},
		{
			name: "type match",
			exp:  true,
			matchConfig: &config.Match{
				Events: []*config.EventType{
					{
						Name: "push",
					},
				},
			},
			event: &domain.Event{
				Type:    "push",
				Payload: &domain.Payload{},
			},
		},
		{
			name: "action match",
			exp:  true,
			matchConfig: &config.Match{
				Events: []*config.EventType{
					{
						Name: "pull_request",
						// OR condition
						Types: []string{"closed", "opened"},
					},
				},
			},
			event: &domain.Event{
				Type: "pull_request",
				Payload: &domain.Payload{
					Action: "opened",
				},
			},
		},
		{
			name: "ignore push delete",
			matchConfig: &config.Match{
				Events: []*config.EventType{
					{
						Name: "push",
					},
				},
			},
			event: &domain.Event{
				Type: "push",
				Payload: &domain.Payload{
					Deleted: true,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, _, err := matchEventType(ctx, tt.matchConfig, tt.event)
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
