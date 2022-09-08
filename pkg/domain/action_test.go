package domain_test

import (
	"context"
	"sort"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"github.com/google/go-cmp/cmp"
)

func TestEvent_GetChangedFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ev      *domain.Event
		wantErr bool
		exp     []string
	}{
		{
			name: "already set",
			ev: &domain.Event{
				ChangedFileObjs: []*github.CommitFile{
					{
						Filename: util.StrP("foo"),
					},
				},
				ChangedFiles: []string{"foo"},
			},
			exp: []string{"foo"},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files, err := tt.ev.GetChangedFiles(ctx)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			sort.Strings(files)
			if diff := cmp.Diff(files, tt.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
