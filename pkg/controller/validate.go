package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"go.uber.org/zap"
)

func (ctrl *Controller) validate(logger *zap.Logger, req *domain.Request) (*githubapp.GitHubApp, *domain.Event, *domain.Response) {
	headers := req.Params.Headers
	bodyStr := req.Body
	appIDS, ok := headers["X-GITHUB-HOOK-INSTALLATION-TARGET-ID"]
	if !ok {
		logger.Warn("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID is required")
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-GITHUB-HOOK-INSTALLATION-TARGET-ID is required",
			},
		}
	}
	appID, err := util.ParseInt64(appIDS)
	if err != nil {
		logger.Warn("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID must be integer")
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-GITHUB-HOOK-INSTALLATION-TARGET-ID must be integer",
			},
		}
	}
	ghApp, ok := ctrl.ghs[appID]
	if !ok {
		logger.Warn("unknown GitHub App ID", zap.Int64("github_app_id", appID))
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "unknown GitHub App ID",
			},
		}
	}

	sig, ok := headers["X-HUB-SIGNATURE"]
	if !ok {
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header X-HUB-SIGNATURE is required",
			},
		}
	}

	bodyB := []byte(bodyStr)
	if err := github.ValidateSignature(sig, bodyB, []byte(ghApp.WebhookSecret)); err != nil {
		logger.Warn("validate the webhook signature", zap.Error(err))
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}
	}

	evType, ok := headers["X-GITHUB-EVENT"]
	if !ok {
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "header x-github-event is required",
			},
		}
	}

	body, err := github.ParseWebHook(evType, bodyB)
	if err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}
	}

	raw := map[string]interface{}{}
	if err := json.Unmarshal(bodyB, &raw); err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}
	}

	payload := &domain.Payload{}
	if err := json.Unmarshal(bodyB, payload); err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, &domain.Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}
	}

	return ghApp, &domain.Event{
		Body:    body,
		Raw:     raw,
		Type:    evType,
		Request: req,
		GitHub:  ghApp.Client,
		Payload: payload,
	}, nil
}
