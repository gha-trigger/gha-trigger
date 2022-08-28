package aws

import (
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

type (
	GetSecretValueInput  = secretsmanager.GetSecretValueInput
	GetSecretValueOutput = secretsmanager.GetSecretValueOutput
	Option               = request.Option
)

func String(s string) *string {
	return &s
}

func (cl *Client) GetSecretValueWithContext(ctx aws.Context, input *GetSecretValueInput, opts ...Option) (*GetSecretValueOutput, error) {
	return cl.secretsManager.GetSecretValueWithContext(ctx, input, opts...)
}
