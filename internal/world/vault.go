package world

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

func (w *World) Vault() *Vault {
	if w.vault != nil {
		return w.vault
	}
	var client *vault.Client
	var err error
	var token string
	vaultConfig := &vault.Config{}
	if err := vaultConfig.ReadEnvironment(); err != nil {
		if w.logger != nil {
			w.logger.Warnf("Failed to read Vault configuration: %s", err.Error())
		}
	}
	client, err = vault.NewClient(vaultConfig)
	if err == nil {
		token = os.Getenv("VAULT_TOKEN")
		if w.logger != nil && token == "" {
			w.logger.Warnf("VAULT_TOKEN not set. Vault not available.")
		} else {
			client.SetToken(token)
		}
	} else {
		if w.logger != nil && token == "" {
			w.logger.Warnf("Failed to create Vault client: %s", err.Error())
		}
	}

	w.vault = &Vault{
		client:     client,
		err:        err,
		logger:     w.logger,
		KeyMapping: make(map[string]string),
	}
	return w.vault
}

type Vault struct {
	logger     *logrus.Logger
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
