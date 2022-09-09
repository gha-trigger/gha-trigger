package lambda

import (
	"context"
	"testing"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"go.uber.org/zap"
)

type mockController struct {
	err error
}

func (ctrl *mockController) Do(ctx context.Context, logger *zap.Logger, req *domain.Request) error {
	return ctrl.err
}

func TestHandler_Do(t *testing.T) {
	t.Parallel()
	logger, _ := zap.NewProduction()
	tests := []struct {
		name    string
		handler *Handler
		wantErr bool
		req     *domain.Request
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
			err := tt.handler.Do(ctx, tt.req)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
