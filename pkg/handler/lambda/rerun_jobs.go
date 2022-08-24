package lambda

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func (handler *Handler) rerunJobs(ctx context.Context, logger *zap.Logger, owner, repo, cmtBody string) (*Response, error) {
	// /rerun-job <job id> [<job id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")
	if len(words) < 2 { //nolint:gomnd
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "job ids are required",
			},
		}, nil
	}

	var gErr error
	resp := &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "jobs are rerun",
		},
	}
	for _, jobID := range words[1:] {
		runID, err := parseInt64(jobID)
		if err != nil {
			logger.Warn("parse a job id as int64", zap.Error(err))
			if resp.StatusCode == http.StatusOK {
				resp.StatusCode = http.StatusBadRequest
			}
			continue
		}
		if res, err := handler.gh.RerunJob(ctx, owner, repo, runID); err != nil {
			logger.Error("rerun a job", zap.Error(err), zap.Int("status_code", res.StatusCode))
			resp = &Response{
				StatusCode: http.StatusInternalServerError,
				Body: map[string]interface{}{
					"message": "failed to rerun a job",
				},
			}
			continue
		}
	}
	return resp, gErr
}
