package secret

import (
	"context"
)

type Provider string

const (
	ProviderAWS Provider = "aws"
)

type SecretManager interface {
	GetSecret(ctx context.Context, secretName string) (string, error)
}
