package util_test

import (
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/util"
)

func TestParseInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wantErr bool
		s       string
		exp     int64
	}{
		{
			name: "normal",
			s:    "10",
			exp:  10,
		},
		{
			name:    "not int",
			s:       "foo",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			i, err := util.ParseInt64(tt.s)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if i != tt.exp {
				t.Fatalf("wanted %v, got %v", tt.exp, i)
			}
		})
	}
}
