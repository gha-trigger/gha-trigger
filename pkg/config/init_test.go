package config

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
		cfg     *Config
		exp     *Config
	}{
		{
			name: "normal",
			cfg: &Config{
				Repos: []*Repo{
					{
						Events: []*Event{
							{
								Matches: []*Match{
									{
										Events: []*EventType{
											{
												Name: "pull_request",
											},
											{
												Name: "push",
											},
										},
									},
									{
										Branches: []*StringMatch{
											{
												Type:  "regexp",
												Value: "^pr/",
											},
											{
												Type:  "equal",
												Value: "main",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			exp: &Config{
				Repos: []*Repo{
					{
						Events: []*Event{
							{
								Matches: []*Match{
									{
										Events: []*EventType{
											{
												Name:  "pull_request",
												Types: []string{"opened", "synchronize", "reopened"},
											},
											{
												Name: "push",
											},
										},
									},
									{
										Branches: []*StringMatch{
											{
												Type:   "regexp",
												Value:  "^pr/",
												regexp: regexp.MustCompile("^pr/"),
											},
											{
												Type:  "equal",
												Value: "main",
											},
										},
									},
								},
							},
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
			err := Init(tt.cfg)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			opt := cmp.AllowUnexported(StringMatch{}, regexp.Regexp{})
			if diff := cmp.Diff(tt.cfg, tt.exp, opt); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
