package lambda

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

type mockController struct {
	resp *domain.Response
	err  error
}

func (ctrl *mockController) Do(ctx context.Context, logger *zap.Logger, req *domain.Request) (*domain.Response, error) {
	return ctrl.resp, ctrl.err
}

func TestHandler_Do(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	tests := []struct {
		name    string
		handler *Handler
		wantErr bool
		req     *domain.Request
		resp    *domain.Response
	}{
		{
			name: "normal",
			handler: &Handler{
				logger: logger,
				ctrl:   &mockController{},
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp, err := tt.handler.Do(ctx, tt.req)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(resp, tt.resp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
