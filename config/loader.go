package config

import (
	"context"
	"fmt"
	"reflect"

	"github.com/NusaCrew/atlas-go/log"
	"github.com/NusaCrew/atlas-go/secret"
	secret_factory "github.com/NusaCrew/atlas-go/secret/factory"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config interface {
	IsEnableLoadingSecret() bool
	GetSecretManagerName() string
	GetSecretManagerReferenceID() string
}

func LoadConfig[T Config](ctx context.Context, cfg T) error {
	if err := godotenv.Load(); err != nil {
		log.Error("%s", err.Error())
	}

	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	if !cfg.IsEnableLoadingSecret() {
		return nil
	}

	secretManager, err := secret_factory.NewSecretManager(ctx, secret.Provider(cfg.GetSecretManagerName()), cfg.GetSecretManagerReferenceID())
	if err != nil {
		return err
	}

	err = loadSecretToConfig(ctx, secretManager, cfg)
	if err != nil {
		return err
	}

	return nil
}

func loadSecretToConfig(ctx context.Context, manager secret.SecretManager, cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be a pointer to struct")
	}

	return populateSecrets(ctx, manager, v.Elem())
}

func populateSecrets(ctx context.Context, manager secret.SecretManager, v reflect.Value) error {
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		val := v.Field(i)

		if !val.CanSet() {
			continue
		}

		if val.Kind() == reflect.Struct {
			if err := populateSecrets(ctx, manager, val); err != nil {
				return err
			}
			continue
		}

		if secretName := field.Tag.Get("secret"); secretName != "" {
			secretVal, err := manager.GetSecret(ctx, secretName)
			if err != nil {
				return fmt.Errorf("failed to get secret for field %s: %w", field.Name, err)
			}

			if val.Kind() == reflect.String {
				val.SetString(secretVal)
			} else {
				return fmt.Errorf("field %s must be a string to hold secret", field.Name)
			}
		}
	}

	return nil
}
