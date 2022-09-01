package lambda

import (
	"context"
	"net/http"
	"strings"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) rerunWorkflows(ctx context.Context, logger *zap.Logger, gh *github.Client, owner, repo, cmtBody string) (*Response, error) {
	// /rerun-workflow <workflow id> [<workflow id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")
	if len(words) < 2 { //nolint:gomnd
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}, nil
	}
	var gErr error
	resp := &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "workflows are rerun",
		},
	}
	for _, workflowRunID := range words[1:] {
		// TODO validation
		runID, err := parseInt64(workflowRunID)
		if err != nil {
			logger.Warn("parse a workflow run id as int64", zap.Error(err))
			if resp.StatusCode == http.StatusOK {
				resp.StatusCode = http.StatusBadRequest
			}
			continue
		}
		logger := logger.With(zap.Int64("workflow_run_id", runID))
		logger.Info("rerunning a workflow")
		if res, err := gh.RerunWorkflow(ctx, owner, repo, runID); err != nil {
			logger.Error("rerun a workflow", zap.Error(err), zap.Int("status_code", res.StatusCode))
			resp = &Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"message": "failed to rerun a workflow",
				},
			}
			continue
		}
	}
	return resp, gErr
}
