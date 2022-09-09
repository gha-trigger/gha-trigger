package lambda

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"go.uber.org/zap"
)

func (handler *Handler) Do(ctx context.Context, req *domain.Request) error {
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

	err := handler.ctrl.Do(ctx, logger, req)
	if util.IsWarn(err) {
		logger.Warn("handle a request", zap.Error(err))
		return nil
	}
	return err
}
