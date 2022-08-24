package lambda

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func (handler *Handler) cancelWorkflows(ctx context.Context, logger *zap.Logger, owner, repo, cmtBody string) (*Response, error) {
	// /cancel <workflow id> [<workflow id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")
	if len(words) < 2 { //nolint:gomnd
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "workflow ids are required",
			},
		}, nil
	}
	var gErr error
	resp := &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "workflows are cancelled",
		},
	}
	for _, workflowID := range words[1:] {
		runID, err := parseInt64(workflowID)
		if err != nil {
			logger.Warn("parse a workflow run id as int64", zap.Error(err))
			if resp.StatusCode == http.StatusOK {
				resp.StatusCode = http.StatusBadRequest
			}
			continue
		}
		if res, err := handler.gh.CancelWorkflow(ctx, owner, repo, runID); err != nil {
			logger.Error("cancel a workflow", zap.Error(err), zap.Int("status_code", res.StatusCode))
			resp = &Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"message": "failed to cancel a workflow",
				},
			}
			continue
		}
	}
	return resp, gErr
}
