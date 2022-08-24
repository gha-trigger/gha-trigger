package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
)

type Client struct {
	secretsManager SecretsManager
}

type SecretsManager interface {
	GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

func New(cfg *config.AWS) *Client {
	sess := session.Must(session.NewSession())
	return &Client{
		secretsManager: secretsmanager.New(sess, aws.NewConfig().WithRegion(cfg.Region)),
	}
}

func (cl *Client) GetSecret(ctx context.Context, param *config.SecretsManager) (*config.Secret, error) {
	ret := &config.Secret{}
	svc := cl.secretsManager
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
