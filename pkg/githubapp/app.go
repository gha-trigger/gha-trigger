package githubapp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/aws"
	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/github"
	"github.com/gha-trigger/gha-trigger/pkg/util"
)

type GitHubApp struct {
	Name          string
	WebhookSecret string
	Client        *github.Client
}

func New(ctx context.Context, awsClient *aws.Client, appCfg *config.GitHubApp) (*GitHubApp, error) {
	paramNewApp := &github.ParamNewApp{
		AppID:          appCfg.AppID,
		InstallationID: appCfg.InstallationID,
		Org:            appCfg.Org,
		User:           appCfg.User,
	}
	input := &aws.GetSecretValueInput{
		SecretId: util.StrP(appCfg.Secret.SecretID),
	}
	if appCfg.Secret.VersionID != "" {
		input.VersionId = util.StrP(appCfg.Secret.VersionID)
	}
	secretOutput, err := awsClient.GetSecretValueWithContext(ctx, input) //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("read the secret value from AWS Secrets Manager: %w", err)
	}
	secret := &config.GitHubAppSecret{}
	if err := json.Unmarshal([]byte(*secretOutput.SecretString), secret); err != nil {
		return nil, fmt.Errorf("unmarshal the GitHub App Secret as JSON: %w", err)
	}
	paramNewApp.KeyFile = secret.GitHubAppPrivateKey
	if secret.AppID != 0 {
		paramNewApp.AppID = secret.AppID
	}
	if secret.InstallationID != 0 {
		paramNewApp.InstallationID = secret.InstallationID
	}
	gh, err := github.NewApp(ctx, paramNewApp)
	if err != nil {
		return nil, fmt.Errorf("create a GitHub Client: %w", err)
	}
	return &GitHubApp{
		Name:          appCfg.Name,
		WebhookSecret: secret.WebhookSecret,
		Client:        gh,
	}, nil
}
