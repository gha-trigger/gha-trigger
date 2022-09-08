package domain

import (
	"sort"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"github.com/google/go-cmp/cmp"
)

func Test_getChangedFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		files []*github.CommitFile
		exp   []string
	}{
		{
			name: "normal",
			files: []*github.CommitFile{
				{
					Filename: util.StrP("foo"),
				},
				{
					Filename:         util.StrP("bar"),
					PreviousFilename: util.StrP("zoo"),
				},
			},
			exp: []string{
				"bar",
				"foo",
				"zoo",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := getChangedFiles(tt.files)
			sort.Strings(files)
			if diff := cmp.Diff(files, tt.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
