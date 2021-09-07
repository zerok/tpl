package world

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	vault "github.com/hashicorp/vault/api"
)

func (w *World) Vault() *Vault {
	if w.vault != nil {
		return w.vault
	}
	logger := zerolog.Ctx(w.ctx).With().Str("component", "Vault").Logger()
	ctx := logger.WithContext(w.ctx)
	var client *vault.Client
	var err error
	var token string
	vaultConfig := &vault.Config{}
	if err := vaultConfig.ReadEnvironment(); err != nil {
		logger.Warn().Msgf("Failed to read Vault configuration: %s", err.Error())
	}
	client, err = vault.NewClient(vaultConfig)
	if err == nil {
		token = os.Getenv("VAULT_TOKEN")
		if token == "" {
			logger.Warn().Msgf("VAULT_TOKEN not set. Vault not available.")
		} else {
			client.SetToken(token)
		}
	} else {
		if token == "" {
			logger.Warn().Msgf("Failed to create Vault client: %s", err.Error())
		}
	}

	w.vault = &Vault{
		ctx:        ctx,
		client:     client,
		err:        err,
		KeyMapping: make(map[string]string),
	}
	return w.vault
}

type Vault struct {
	ctx        context.Context
	client     *vault.Client
	err        error
	Prefix     string
	KeyMapping map[string]string
}

func (v *Vault) Secret(path, field string) (string, error) {
	if v.client == nil {
		return "", errors.New("no vault client available")
	}
	prefixPath := fmt.Sprintf("%s%s", v.Prefix, path)
	mapped, ok := v.KeyMapping[prefixPath]
	if !ok {
		mapped = path
	}
	sec, err := v.client.Logical().Read(mapped)
	if err != nil {
		return "", errors.Wrapf(err, "failed to access Vault path %s", mapped)
	}
	if sec == nil {
		return "", errors.Errorf("Vault path %s contained no secret", mapped)
	}
	raw, ok := sec.Data[field]
	if !ok {
		return "", errors.Errorf("%s has no field named '%s'", mapped, field)
	}
	return fmt.Sprintf("%s", raw), nil
}
