package controller

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/domain"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/githubapp"
	"github.com/gha-trigger/gha-trigger/pkg/util"
	"github.com/suzuki-shunsuke/zap-error/logerr"
	"go.uber.org/zap"
)

var (
	errHeaderXGitHubHookInstallationTargetIDIsRequred   = util.WithWarn(errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID is required"))
	errHeaderXGitHubHookInstallationTargetIDMustBeInt64 = util.WithWarn(errors.New("header X-GITHUB-HOOK-INSTALLATION-TARGET-ID must be integer"))
	errUnknownGitHubAppID                               = util.WithWarn(errors.New("unknown GitHub App ID"))
	errHeaderXHubSignatureIsRequired                    = util.WithWarn(errors.New("header X-HUB-SIGNATURE is required"))
	errSignatureInvalid                                 = util.WithWarn(errors.New("signature is invalid"))
	errHeaderXHubEventIsRequired                        = util.WithWarn(errors.New("header X-HUB-EVENT is required"))
)

func (ctrl *Controller) validate(logger *zap.Logger, req *domain.Request) (*githubapp.GitHubApp, *domain.Event, error) {
	headers := req.Params.Headers
	bodyStr := req.Body
	appIDS, ok := headers["X-GITHUB-HOOK-INSTALLATION-TARGET-ID"]
	if !ok {
		return nil, nil, errHeaderXGitHubHookInstallationTargetIDIsRequred
	}
	appID, err := util.ParseInt64(appIDS)
	if err != nil {
		return nil, nil, errHeaderXGitHubHookInstallationTargetIDMustBeInt64
	}
	ghApp, ok := ctrl.ghs[appID]
	if !ok {
		return nil, nil, logerr.WithFields(errUnknownGitHubAppID, zap.Int64("github_app_id", appID))
	}

	sig, ok := headers["X-HUB-SIGNATURE"]
	if !ok {
		return nil, nil, errHeaderXHubSignatureIsRequired
	}

	bodyB := []byte(bodyStr)
	if err := github.ValidateSignature(sig, bodyB, []byte(ghApp.WebhookSecret)); err != nil {
		logger.Warn("validate the webhook signature", zap.Error(err))
		return nil, nil, errSignatureInvalid
	}

	evType, ok := headers["X-GITHUB-EVENT"]
	if !ok {
		return nil, nil, errHeaderXHubEventIsRequired
	}

	body, err := github.ParseWebHook(evType, bodyB)
	if err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, fmt.Errorf("parse a webhook payload: %w", err)
	}

	raw := map[string]interface{}{}
	if err := json.Unmarshal(bodyB, &raw); err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, fmt.Errorf("parse a webhook payload: %w", err)
	}

	payload := &domain.Payload{}
	if err := json.Unmarshal(bodyB, payload); err != nil {
		logger.Warn("parse a webhook payload", zap.Error(err))
		return nil, nil, fmt.Errorf("parse a webhook payload: %w", err)
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
