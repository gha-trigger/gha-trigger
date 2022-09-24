package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/gha-trigger/gha-trigger/pkg/config"
)

type Client struct {
	secretsManager SecretsManager
}

type SecretsManager interface {
	GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

func New(cfg *config.AWS) *Client {
	sess := session.Must(session.NewSession())
	awsCfg := aws.NewConfig()
	if region := cfg.GetRegion(); region != "" {
		awsCfg.WithRegion(region)
	}
	return &Client{
		secretsManager: secretsmanager.New(sess, awsCfg),
	}
}

type (
	GetSecretValueInput  = secretsmanager.GetSecretValueInput
	GetSecretValueOutput = secretsmanager.GetSecretValueOutput
	Option               = request.Option
	Context              = aws.Context
)

func (cl *Client) GetSecretValueWithContext(ctx aws.Context, input *GetSecretValueInput, opts ...Option) (*GetSecretValueOutput, error) {
	return cl.secretsManager.GetSecretValueWithContext(ctx, input, opts...)
}
