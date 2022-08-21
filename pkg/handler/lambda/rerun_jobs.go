package lambda

import (
	"context"
	"net/http"
	"strings"
)

func (handler *Handler) rerunJobs(ctx context.Context, event *Event, cmtBody string) (*Response, error) { //nolint:unparam
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
	return &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "jobs are rerun",
		},
	}, nil
}
