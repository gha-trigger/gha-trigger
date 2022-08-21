package lambda

import (
	"context"
	"net/http"
	"strings"
)

func (handler *Handler) rerunWorkflows(ctx context.Context, event *Event, cmtBody string) (*Response, error) { //nolint:unparam
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
	return &Response{
		StatusCode: http.StatusOK,
		Body: map[string]interface{}{
			"message": "workflows are rerun",
		},
	}, nil
}
