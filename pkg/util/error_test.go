package util_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/util"
)

func TestWarnError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		wantWarn bool
		exp      string
		err      error
	}{
		{
			name: "not warn",
			exp:  "foo",
			err:  errors.New("foo"),
		},
		{
			name:     "warn",
			exp:      "foo: bar",
			wantWarn: true,
			err:      fmt.Errorf("foo: %w", util.WithWarn(errors.New("bar"))),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := util.IsWarn(tt.err)
			if tt.wantWarn != f {
				t.Fatalf("wanted %v, got %v", tt.wantWarn, f)
			}
			s := tt.err.Error()
			if s != tt.exp {
				t.Fatalf("wanted %v, got %v", tt.exp, s)
			}
		})
	}
}
