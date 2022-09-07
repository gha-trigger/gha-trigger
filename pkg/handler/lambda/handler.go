package lambda

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func (handler *Handler) Do(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// func (handler *Handler) Do(ctx context.Context, e interface{}) (*Response, error) {
	// 	logger := handler.logger
	// 	logger.Info("start a request")
	// 	defer logger.Info("end a request")
	//
	// 	var req *Request
	//
	// 	if err := json.NewEncoder(os.Stderr).Encode(e); err != nil {
	// 		return nil, err
	// 	}
	logger := handler.logger
	logger.Info("start a request")
	defer logger.Info("end a request")

	return handler.ctrl.Do(ctx, logger, req)
}
