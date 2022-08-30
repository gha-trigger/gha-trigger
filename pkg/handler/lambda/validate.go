package lambda

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
	"go.uber.org/zap"
)

func (handler *Handler) validate(logger *zap.Logger, event *events.APIGatewayProxyRequest) (*GitHubApp, *Event, *Response) {
	appIDS, ok := event.Headers["X-GITHUB-HOOK-INSTALLATION-TARGET-ID"]
	if !ok {
		logger.Warn("header X-GitHub-Hook-Installation-Target-ID is required")
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-GitHub-Hook-Installation-Target-ID is required",
			},
		}
	}
	appID, err := parseInt64(appIDS)
	if err != nil {
		logger.Warn("header X-GitHub-Hook-Installation-Target-ID must be integer")
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-GitHub-Hook-Installation-Target-ID must be integer",
			},
		}
	}
	ghApp, ok := handler.ghs[appID]
	if !ok {
		logger.Warn("unknown GitHub App ID", zap.Int64("github_app_id", appID))
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "unknown GitHub App ID",
			},
		}
	}

	sig, ok := event.Headers["X-HUB-SIGNATURE"]
	if !ok {
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-Hub-Signature is required",
			},
		}
	}

	if err := github.ValidateSignature(sig, []byte(event.Body), []byte(ghApp.WebhookSecret)); err != nil {
		logger.Warn("validate the webhook signature", zap.Error(err))
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}
	}

	evType, ok := event.Headers["X-GITHUB-EVENT"]
	if !ok {
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header x-github-event is required",
			},
		}
	}

	body, err := github.ParseWebHook(evType, []byte(event.Body))
	if err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}
	}
	return ghApp, &Event{
		Body:  body,
		Type:  evType,
		Event: event,
	}, nil
}
