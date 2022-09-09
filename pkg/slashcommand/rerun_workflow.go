package slashcommand

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type WorkflowRerunner interface {
	RerunWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
}

func rerunWorkflows(ctx context.Context, logger *zap.Logger, gh WorkflowRerunner, owner, repo string, words []string) {
	// /rerun-workflow <workflow id> [<workflow id> ...]
	if len(words) < 2 { //nolint:gomnd
		// TODO send notification to issue or pr
		return
	}

	ids, err := parseIDs(words[1:])
	if err != nil {
		logger.Warn("parse a workflow run id as int64", zap.Error(err))
		return
	}

	for _, runID := range ids {
		logger := logger.With(zap.Int64("workflow_run_id", runID))
		logger.Info("rerunning a workflow")
		if res, err := gh.RerunWorkflow(ctx, owner, repo, runID); err != nil {
			// TODO send a notification to pr or issue
			logger.Error("rerun a workflow", zap.Error(err), zap.Int("status_code", res.StatusCode))
		}
	}
}
