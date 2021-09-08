package world

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	AzureTenantId               string = "AZURE_TENANT_ID"
	AzureClientId               string = "AZURE_CLIENT_ID"
	AzureClientSecret           string = "AZURE_CLIENT_SECRET"
	AzureKeyVaultUrl            string = "AZURE_KEY_VAULT_URL"
	AzureApiVersion             string = "AZURE_API_VERSION"
	AzureToken                  string = "AZURE_TOKEN"
	AzureVaultUrl               string = "https://vault.azure.net"
	AzureClientCredentialsGrant string = "client_credentials"
	MicrosoftLoginUrl           string = "https://login.microsoftonline.com/"
)

type AzureOauth2Res struct {
	AccessToken string `json:"access_token"`
}

type AzureKeyVaultEntry struct {
	Value string `json:"value"`
}

type AzureSecretVersions struct {
	Value []struct {
		ID         string `json:"id"`
		Attributes struct {
			Enabled       bool   `json:"enabled"`
			Created       int    `json:"created"`
			Updated       int    `json:"updated"`
			RecoveryLevel string `json:"recoveryLevel"`
		} `json:"attributes"`
	} `json:"value"`
}

type Azure struct {
	ctx          context.Context
	Prefix       string
	KeyMapping   map[string]string
	keyVaultUrl  string
	tenantId     string
	clientId     string
	clientSecret string
	apiVersion   string
	token        string
}

// LeveledZerolog implements the retryablehttp LeveledLogger interface
// to use zerolog as logging backend. See:
// https://github.com/hashicorp/go-retryablehttp/pull/101#issuecomment-735206810
type LeveledZerolog struct {
	logger *zerolog.Logger
}

func (l *LeveledZerolog) fields(keysAndValues ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})

	for i := 0; i < len(keysAndValues)-1; i += 2 {
		fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}

	return fields
}

func (l *LeveledZerolog) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error().Fields(l.fields(keysAndValues...)).Msg(msg)
}

func (l *LeveledZerolog) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info().Fields(l.fields(keysAndValues...)).Msg(msg)
}
func (l *LeveledZerolog) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug().Fields(l.fields(keysAndValues...)).Msg(msg)
}

func (l *LeveledZerolog) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn().Fields(l.fields(keysAndValues...)).Msg(msg)
}

func (w *World) Azure() *Azure {
	if w.azure != nil {
		return w.azure
	}
	logger := zerolog.Ctx(w.ctx).With().Str("component", "Azure").Logger()
	ctx := logger.WithContext(w.ctx)
	tenantId := os.Getenv(AzureTenantId)
	azureClientId := os.Getenv(AzureClientId)
	azureClientSecret := os.Getenv(AzureClientSecret)
	azureKeyVaultUrl := os.Getenv(AzureKeyVaultUrl)
	azureApiVersion := os.Getenv(AzureApiVersion)
	azureToken := os.Getenv(AzureToken)

	if azureApiVersion == "" {
		azureApiVersion = "7.0"
	}

	if azureKeyVaultUrl == "" {
		logger.Warn().Msgf("%v not set.", AzureKeyVaultUrl)
	}

	if azureToken == "" && (tenantId == "" || azureClientId == "" || azureClientSecret == "") {
		logger.Warn().Msgf("%s or %s, %s, %s needs to be set", AzureToken, AzureTenantId, AzureClientId, AzureClientSecret)
	}

	w.azure = &Azure{
		ctx:          ctx,
		KeyMapping:   make(map[string]string),
		tenantId:     tenantId,
		clientId:     azureClientId,
		clientSecret: azureClientSecret,
		keyVaultUrl:  azureKeyVaultUrl,
		apiVersion:   azureApiVersion,
		token:        azureToken,
	}
	return w.azure
}

func (a *Azure) Secret(path string) (string, error) {
	prefixPath := fmt.Sprintf("%s%s", a.Prefix, path)
	mapped, ok := a.KeyMapping[prefixPath]
	if !ok {
		mapped = path
	}
	latestSecretVersion, err := a.getLatestSecretVersion(mapped)
	if err != nil {
		return "", errors.Wrapf(err, "could not get secrets version for %s", mapped)
	}
	secret, err := a.getSecret(mapped, latestSecretVersion)
	if err != nil {
		return "", errors.Wrapf(err, "could not get secrets for %s", mapped)
	}
	return secret, nil
}

func (a *Azure) getSecret(path string, secretVersion string) (string, error) {
	body, err := a.doVaultRequest(fmt.Sprintf("/secrets/%s/%s", path, secretVersion))
	if err != nil {
		return "", err
	}
	var entry AzureKeyVaultEntry
	err = json.Unmarshal(body, &entry)
	if err != nil {
		return "", err
	}
	return entry.Value, nil
}

func (a *Azure) getLatestSecretVersion(path string) (string, error) {
	body, err := a.doVaultRequest(fmt.Sprintf("/secrets/%s/versions", path))
	if err != nil {
		return "", err
	}
	var secretVersions AzureSecretVersions
	err = json.Unmarshal(body, &secretVersions)
	if err != nil {
		return "", err
	}
	latestVersion := secretVersions.Value[0]
	for _, v := range secretVersions.Value {
		if v.Attributes.Created > latestVersion.Attributes.Created {
			latestVersion = v
		}
	}
	// version value is returned as an URL so we split it and return the last part which contains the version string
	split := strings.Split(latestVersion.ID, "/")
	return split[len(split)-1], nil
}

func (a *Azure) doVaultRequest(urlPath string) ([]byte, error) {
	logger := zerolog.Ctx(a.ctx)
	if a.token == "" {
		if err := a.getBearerToken(); err != nil {
			return nil, errors.Wrap(err, "failed to retrieve token")
		}
	}
	params := url.Values{}
	params.Set("api-version", a.apiVersion)
	u, err := url.ParseRequestURI(a.keyVaultUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vault request URI")
	}
	u.Path = urlPath
	u.RawQuery = params.Encode()

	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate request")
	}
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.token))

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = &LeveledZerolog{logger}
	client := retryClient.StandardClient()

	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("request returned error, code: %v", resp.StatusCode))
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (a *Azure) getBearerToken() error {
	logger := zerolog.Ctx(a.ctx)
	params := url.Values{}
	params.Set("grant_type", AzureClientCredentialsGrant)
	params.Set("client_id", a.clientId)
	params.Set("client_secret", a.clientSecret)
	params.Set("resource", AzureVaultUrl)
	u, err := url.ParseRequestURI(MicrosoftLoginUrl)
	if err != nil {
		return errors.Wrap(err, "failed to parse login URL")
	}
	u.Path = fmt.Sprintf("/%s/oauth2/token", a.tenantId)

	r, err := http.NewRequest("POST", u.String(), strings.NewReader(params.Encode()))
	if err != nil {
		return errors.Wrap(err, "request generation failed for token")
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = &LeveledZerolog{logger}
	client := retryClient.StandardClient()

	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrap(err, "token request failed")
	}
	var bearerTokenResponse AzureOauth2Res
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &bearerTokenResponse)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal token response")
	}
	a.token = bearerTokenResponse.AccessToken
	return nil
}
