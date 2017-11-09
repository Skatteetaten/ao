package client

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
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

func (p Permissions) AddGroup(group string) error {
	groups := p["groups"]
	for _, g := range groups {
		if g == group {
			return errors.New("Group already exists, " + group)
		}
	}
	p["groups"] = append(groups, group)

	return nil
}

func (p Permissions) DeleteGroup(group string) error {
	groups := p["groups"]
	var hasDeleted bool
	for i, g := range groups {
		if g == group {
			p["groups"] = append(groups[:i], groups[i+1:]...)
			hasDeleted = true
			break
		}
	}
	if !hasDeleted {
		return errors.New("Did not find group " + group)
	}
	return nil
}

func (p Permissions) GetGroups() []string {
	return p["groups"]
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

func (api *ApiClient) DeleteVault(vaultName string) error {
	endpoint := fmt.Sprintf("/affiliation/%s/vault/%s", api.Affiliation, vaultName)

	response, err := api.Do(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}
