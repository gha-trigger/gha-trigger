package lambda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/config"
)

func readSecretFromSecretsManager(ctx context.Context, svc *secretsmanager.SecretsManager, param *config.SecretsManager) (*config.Secret, error) {
	ret := &config.Secret{}
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(param.SecretID),
	}
	if param.VersionID != "" {
		input.VersionId = aws.String(param.VersionID)
	}
	output, err := svc.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return ret, fmt.Errorf("get secret value from AWS SecretsManager: %w", err)
	}
	if err := json.Unmarshal([]byte(*output.SecretString), ret); err != nil {
		return ret, fmt.Errorf("parse secret value: %w", err)
	}
	return ret, nil
}
