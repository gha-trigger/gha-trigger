package slashcommand

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type JobRerunner interface {
	RerunJob(ctx context.Context, owner, repo string, jobID int64) (*github.Response, error)
}

func rerunJobs(ctx context.Context, logger *zap.Logger, gh JobRerunner, owner, repo string, words []string) {
	// /rerun-job <job id> [<job id> ...]
	if len(words) == 0 { //nolint:gomnd
		// TODO send notification to issue or pr
		logger.Warn("job id is required for /rerun-job")
		return
	}

	ids, err := parseIDs(words)
	if err != nil {
		logger.Warn("parse a job id as int64", zap.Error(err))
		return
	}

	for _, jobID := range ids {
		logger := logger.With(zap.Int64("job_id", jobID))
		if res, err := gh.RerunJob(ctx, owner, repo, jobID); err != nil {
			// TODO send a notification to pr or issue
			logger.Error("rerun a job", zap.Error(err), zap.Int("status_code", res.StatusCode))
		}
	}
}
