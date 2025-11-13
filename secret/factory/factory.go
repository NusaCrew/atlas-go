package factory

import (
	"context"
	"fmt"

	"github.com/NusaCrew/atlas-go/secret"
	aws_secret_manager "github.com/NusaCrew/atlas-go/secret/factory/aws"
)

func NewSecretManager(ctx context.Context, provider secret.Provider, referenceID string) (secret.SecretManager, error) {
	switch provider {
	case secret.ProviderAWS:
		return aws_secret_manager.NewAWSSecretManager(ctx, referenceID)
	default:
		return nil, fmt.Errorf("unsupported secret manager: %s", provider)
	}
}
