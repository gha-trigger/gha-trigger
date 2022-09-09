package slashcommand

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

type WorkflowCanceler interface {
	CancelWorkflow(ctx context.Context, owner, repo string, runID int64) (*github.Response, error)
}

// return true if request matches
func cancelWorkflows(ctx context.Context, logger *zap.Logger, gh WorkflowCanceler, owner, repo, cmtBody string) (bool, error) {
	// /cancel <workflow id> [<workflow id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")

	if words[0] != "/cancel" {
		return false, nil
	}

	if len(words) < 2 { //nolint:gomnd
		// TODO send notification to issue or pr
		return true, nil
	}

	ids, err := parseIDs(words[1:])
	if err != nil {
		logger.Warn("parse a workflow run id as int64", zap.Error(err))
		return true, nil
	}

	for _, runID := range ids {
		logger := logger.With(zap.Int64("workflow_run_id", runID))
		logger.Info("cancelling a workflow")
		if res, err := gh.CancelWorkflow(ctx, owner, repo, runID); err != nil {
			// TODO send a notification to pr or issue
			logger.Error("cancel a workflow", zap.Error(err), zap.Int("status_code", res.StatusCode))
		}
	}
	return true, nil
}
