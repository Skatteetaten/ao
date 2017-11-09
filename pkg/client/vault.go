package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type Secrets map[string]string

func (s Secrets) ShowContent(name string) string {
	secret := s[name]
	if secret == "" {
		return ""
	}
	data, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return ""
	}

	return string(data)
}

type Permissions map[string][]string

func (p Permissions) GetGroups() []string {
	permissions, found := p["groups"]
	if !found {
		return []string{}
	}
	return permissions
}

type AuroraSecretVault struct {
	Name        string      `json:"name"`
	Secrets     Secrets     `json:"secrets"`
	Permissions Permissions `json:"permissions"`
}

func (api *ApiClient) GetVaults() ([]*AuroraSecretVault, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/vault", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var vaults []*AuroraSecretVault
	err = response.ParseItems(&vaults)
	if err != nil {
		return nil, err
	}

	return vaults, nil
}

func (api *ApiClient) GetVault(vaultName string) (*AuroraSecretVault, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/vault/%s", api.Affiliation, vaultName)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var vault AuroraSecretVault
	err = response.ParseFirstItem(&vault)
	if err != nil {
		return nil, err
	}

	return &vault, nil
}
