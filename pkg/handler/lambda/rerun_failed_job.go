package lambda

import (
	"context"
	"net/http"
	"strings"
)

func (handler *Handler) rerunFailedJobs(ctx context.Context, event *Event, cmtBody string) (*Response, error) { //nolint:unparam
	// /rerun-failed-job <workflow id> [<workflow id> ...]
	words := strings.Split(strings.TrimSpace(cmtBody), " ")
	if len(words) < 2 { //nolint:gomnd
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "workflow ids are required",
			},
		}, nil
	}
	return &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "failed jobs are rerun",
		},
	}, nil
}
