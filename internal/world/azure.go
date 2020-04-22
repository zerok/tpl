package world

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	AzureSubscriptionId string = "AZURE_SUBSCRIPTION_ID"
	AzureClientId       string = "AZURE_CLIENT_ID"
	AzureClientSecret   string = "AZURE_CLIENT_SECRET"
	AzureKeyVaultUrl    string = "AZURE_KEY_VAULT_URL"
	AzureApiVersion     string = "AZURE_API_VERSION"
)

type Oauth2Res struct {
	AccessToken string `json:"access_token"`
}

type KeyVaultEntry struct {
	Value string `json:"value"`
}

type SecretVersions struct {
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
	logger         *logrus.Logger
	Prefix         string
	KeyMapping     map[string]string
	keyVaultUrl    string
	subscriptionId string
	clientId       string
	clientSecret   string
	apiVersion     string
	token          string
}

func (w *World) Azure() *Azure {
	if w.azure != nil {
		return w.azure
	}
	azureSubscriptionId := w.checkAzureEnv(AzureSubscriptionId)
	azureClientId       := w.checkAzureEnv(AzureClientId)
	azureClientSecret   := w.checkAzureEnv(AzureClientSecret)
	azureKeyVaultUrl    := w.checkAzureEnv(AzureKeyVaultUrl)
	azureApiVersion     := w.checkAzureEnv(AzureApiVersion)

	if azureApiVersion == "" {
		azureApiVersion = "7.0"
	}

	w.azure = &Azure{
		logger:         w.logger,
		KeyMapping:     make(map[string]string),
		subscriptionId: azureSubscriptionId,
		clientId:       azureClientId,
		clientSecret:   azureClientSecret,
		keyVaultUrl:    azureKeyVaultUrl,
		apiVersion:     azureApiVersion,
	}
	return w.azure
}

func (w *World) checkAzureEnv(env string) string {
	value := os.Getenv(env)
	if w.logger != nil && value == "" {
		w.logger.Warnf("%v not set.", env)
	}
	return value
}

func (a *Azure) Secret(path string) (string, error) {
	prefixPath := fmt.Sprintf("%s%s", a.Prefix, path)
	mapped, ok := a.KeyMapping[prefixPath]
	if !ok {
		mapped = path
	}
	// azure keyvault only allows alphanumeric chars and dashes, for easier use and backwards compatibility we
	// replace "/" with "--" and use the double dashes as a separator
	mapped = strings.Replace(mapped, "/", "--", -1)
	err := a.getBearerToken()
	if err != nil {
		return "", errors.Wrap(err, "could not get access token from https://login.microsoftonline.com/")
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
	var entry KeyVaultEntry
	err = json.Unmarshal(body, &entry)
	if err != nil {
		return "", err
	}
	return entry.Value, nil
}

func (a *Azure) getLatestSecretVersion(path string) (string, error) {
	secretPath := path
	body, err := a.doVaultRequest(fmt.Sprintf("/secrets/%s/versions", secretPath))
	if err != nil {
		return "", err
	}
	var secretVersions SecretVersions
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
	return split[len(split) - 1], nil
}

func (a *Azure) doVaultRequest(urlPath string) ([]byte, error) {
	if a.token == "" {
		return nil, errors.New("bearer token missing")
	}
	params := url.Values{}
	params.Set("api-version", a.apiVersion)
	u, err := url.ParseRequestURI(a.keyVaultUrl)
	if err != nil {
		return nil, err
	}
	u.Path = urlPath
	u.RawQuery = params.Encode()
	client := &http.Client{}
	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.token))
	resp, err := client.Do(r)
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("request returned error, code: %v", resp.StatusCode))
	}
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (a *Azure) getBearerToken() error {
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", a.clientId)
	params.Set("client_secret", a.clientSecret)
	params.Set("resource", "https://vault.azure.net")
	u, err := url.ParseRequestURI("https://login.microsoftonline.com/")
	if err != nil {
		return err
	}
	u.Path = fmt.Sprintf("/%s/oauth2/token", a.subscriptionId)
	client := &http.Client{}
	r, err := http.NewRequest("POST", u.String(), strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	var bearerTokenResponse Oauth2Res
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &bearerTokenResponse)
	if err != nil {
		return nil
	}
	a.token = bearerTokenResponse.AccessToken
	return nil
}
