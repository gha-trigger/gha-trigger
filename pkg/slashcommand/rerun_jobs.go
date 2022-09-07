package slashcommand

// import (
// 	"context"
// 	"net/http"
// 	"strings"
//
// 	"github.com/gha-trigger/gha-trigger/pkg/github"
// 	"github.com/gha-trigger/gha-trigger/pkg/util"
// 	"go.uber.org/zap"
// )
//
// func rerunJobs(ctx context.Context, logger *zap.Logger, gh *github.Client, owner, repo, cmtBody string) (*Response, error) {
// 	// /rerun-job <job id> [<job id> ...]
// 	words := strings.Split(strings.TrimSpace(cmtBody), " ")
// 	if len(words) < 2 { //nolint:gomnd
// 		return &Response{
// 			StatusCode: http.StatusBadRequest,
// 			Body: map[string]interface{}{
// 				"error": "job ids are required",
// 			},
// 		}, nil
// 	}
//
// 	var gErr error
// 	resp := &Response{
// 		StatusCode: http.StatusOK,
// 		Body: map[string]interface{}{
// 			"message": "jobs are rerun",
// 		},
// 	}
// 	for _, jobID := range words[1:] {
// 		runID, err := util.ParseInt64(jobID)
// 		if err != nil {
// 			logger.Warn("parse a job id as int64", zap.Error(err))
// 			if resp.StatusCode == http.StatusOK {
// 				resp.StatusCode = http.StatusBadRequest
// 			}
// 			continue
// 		}
// 		if res, err := gh.RerunJob(ctx, owner, repo, runID); err != nil {
// 			logger.Error("rerun a job", zap.Error(err), zap.Int("status_code", res.StatusCode))
// 			resp = &Response{
// 				StatusCode: http.StatusInternalServerError,
// 				Body: map[string]interface{}{
// 					"message": "failed to rerun a job",
// 				},
// 			}
// 			continue
// 		}
// 	}
// 	return resp, gErr
// }
