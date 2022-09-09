package controller

import (
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/google/go-cmp/cmp"
)

func Test_toUpperHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		params  *domain.RequestParamsField
		headers map[string]string
	}{
		{
			name: "normal",
			params: &domain.RequestParamsField{
				Headers: map[string]string{
					"x-github-hook-installation-target-id": "xxx",
				},
			},
			headers: map[string]string{
				"X-GITHUB-HOOK-INSTALLATION-TARGET-ID": "xxx",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			toUpperHeaders(tt.params)
			if diff := cmp.Diff(tt.params.Headers, tt.headers); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
