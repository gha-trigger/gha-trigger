package lambda

import (
	"net/http"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) validate(logger *zap.Logger, event *Event) (interface{}, *Response) {
	if err := github.ValidateSignature(event.Headers.Signature, []byte(event.Body), []byte(handler.secret.WebhookSecret)); err != nil {
		logger.Warn("validate the webhook signature", zap.Error(err))
		return nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}
	}

	body, err := github.ParseWebHook(event.Headers.Event, []byte(event.Body))
	if err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}
	}
	return body, nil
}
