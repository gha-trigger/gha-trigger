package lambda

import (
	"context"
	"net/http"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) rerunFailedJobs(ctx context.Context, logger *zap.Logger, gh *github.Client, owner, repo, cmtBody string) (*domain.Response, error) {
	// /rerun-failed-job <workflow id> [<workflow id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")
	if len(words) < 2 { //nolint:gomnd
		return &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "workflow ids are required",
			},
		}, nil
	}
	var gErr error
	resp := &domain.Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "failed jobs are rerun",
		},
	}
	for _, workflowID := range words[1:] {
		runID, err := parseInt64(workflowID)
		if err != nil {
			logger.Warn("parse a workflow id as int64", zap.Error(err))
			if resp.StatusCode == http.StatusOK {
				resp.StatusCode = http.StatusBadRequest
			}
			continue
		}
		logger := logger.With(zap.Int64("workflow_run_id", runID))
		logger.Info("rerunning failed jobs")
		if res, err := gh.RerunFailedJobs(ctx, owner, repo, runID); err != nil {
			logger.Error(
				"rerun failed jobs", zap.Error(err),
				zap.Int("status_code", res.StatusCode),
			)
			resp = &domain.Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"message": "failed to rerun failed jobs",
				},
			}
			continue
		}
	}
	return resp, gErr
}
