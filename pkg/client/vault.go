package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/graphql"
	"net/http"
	"strings"
)

type (
	// Secrets is a key-value map of secrets
	Secrets map[string]string

	// AuroraVaultInfo TODO: rename to response
	AuroraVaultInfo struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		Secrets     Secrets  `json:"secrets"`
		HasAccess   bool     `json:"hasAccess"`
	}

	// AuroraSecretVault TODO: rename to request
	AuroraSecretVault struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		Secrets     Secrets  `json:"secrets"`
	}

	// VaultFileResource holds contents from a vault file resource
	VaultFileResource struct {
		Contents string `json:"contents"`
	}
)

const ErrorVaultNotFound = "Vault not found"
const FoundNoSecretsForVault = "Found no secrets for vault"

// NewAuroraSecretVault creates a new AuroraSecretVault
func NewAuroraSecretVault(name string) *AuroraSecretVault {
	return &AuroraSecretVault{
		Name:        name,
		Secrets:     make(Secrets),
		Permissions: []string{},
	}
}

// GetVaults gets aurora vault information via API calls
func (api *APIClient) GetVaults() ([]*AuroraVaultInfo, error) {
	endpoint := fmt.Sprintf("/vault/%s", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var vaults []*AuroraVaultInfo
	err = response.ParseItems(&vaults)
	if err != nil {
		return nil, err
	}

	return vaults, nil
}

// GetVault gets an aurora secret vault via API calls
func (api *APIClient) GetVault(vaultName string) (*AuroraSecretVault, error) {
	endpoint := fmt.Sprintf("/vault/%s/%s", api.Affiliation, vaultName)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	if response != nil && !response.Success && strings.Contains(response.Message, "Vault not found") {
		return nil, errors.New(ErrorVaultNotFound)
	}

	var vault AuroraSecretVault
	err = response.ParseFirstItem(&vault)
	if err != nil {
		return nil, err
	}

	return &vault, nil
}

// SaveVault saves an aurora secret vault via API calls
func (api *APIClient) SaveVault(vault AuroraSecretVault) error {
	if len(vault.Permissions) == 0 {
		return errors.New("Aborted: Vault can not be saved without permissions")
	}

	endpoint := fmt.Sprintf("/vault/%s", api.Affiliation)

	data, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	response, err := api.Do(http.MethodPut, endpoint, data)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Message)
	}

	return nil
}

// GetSecretFile gets a secret file via API calls
func (api *APIClient) GetSecretFile(vault, secret string) (string, string, error) {
	endpoint := fmt.Sprintf("/vault/%s/%s/%s", api.Affiliation, vault, secret)

	bundle, err := api.DoWithHeader(http.MethodGet, endpoint, nil, nil)
	if err != nil || bundle == nil {
		return "", "", err
	}

	if !bundle.BooberResponse.Success {
		return "", "", errors.New(bundle.BooberResponse.Message)
	}

	var vaultFile VaultFileResource
	err = bundle.BooberResponse.ParseFirstItem(&vaultFile)
	if err != nil {
		return "", "", nil
	}

	data, err := base64.StdEncoding.DecodeString(vaultFile.Contents)
	if err != nil {
		return "", "", err
	}

	eTag := bundle.HTTPResponse.Header.Get("ETag")

	return string(data), eTag, nil
}

// UpdateSecretFile updates a secret file via API calls
func (api *APIClient) UpdateSecretFile(vault, secret, eTag string, content []byte) error {
	endpoint := fmt.Sprintf("/vault/%s/%s/%s", api.Affiliation, vault, secret)

	encoded := base64.StdEncoding.EncodeToString(content)

	header := map[string]string{
		"If-Match": eTag,
	}

	vaultFile := VaultFileResource{
		Contents: encoded,
	}

	data, err := json.Marshal(vaultFile)
	if err != nil {
		return err
	}

	bundle, err := api.DoWithHeader(http.MethodPut, endpoint, header, data)
	if err != nil || bundle == nil {
		return err
	}

	if !bundle.BooberResponse.Success {
		return errors.New(bundle.BooberResponse.Message)
	}

	return nil
}

// VaultResponse is core of response from the graphql "addVaultPermissions" and "removeVAultPermissions"
type VaultResponse struct {
	HasAccess   bool          `json:"hasAccess"`
	Name        string        `json:"name"`
	Permissions []string      `json:"permissions"`
	Secrets     []interface{} `json:"secrets"`
}

const createVaultRequestString = `mutation createVault($createVaultInput: CreateVaultInput!){
  createVault(input: $createVaultInput)
  {
    hasAccess
    name
    permissions
  }
}`

// CreateVaultInput is input to the graphql createVault interface
type CreateVaultInput struct {
	AffiliationName string        `json:"affiliationName"`
	Permissions     []string      `json:"permissions"`
	Secrets         []SecretInput `json:"secrets"`
	VaultName       string        `json:"vaultName"`
}

// CreateVaultInput is input to the graphql createVault interface
type SecretInput struct {
	Base64Content string `json:"base64Content"`
	Name          string `json:"name"`
}

// AddPermissions adds permissions to vault via gobo
func (api *APIClient) CreateVault(vault AuroraSecretVault) error {
	if len(vault.Secrets) == 0 {
		return errors.New(FoundNoSecretsForVault)
	}

	createVaultRequest := graphql.NewRequest(createVaultRequestString)
	createVaultInput := getCreateVaultInput(vault, api.Affiliation)
	createVaultRequest.Var("createVaultInput", createVaultInput)
	var createVaultResponse VaultResponse
	if err := api.RunGraphQlMutation(createVaultRequest, &createVaultResponse); err != nil {
		return err
	}

	return nil
}

func getCreateVaultInput(vault AuroraSecretVault, affiliation string) CreateVaultInput {
	secrets := make([]SecretInput, len(vault.Secrets))
	i := 0
	for key, content := range vault.Secrets {
		secrets[i] = SecretInput{
			Base64Content: content,
			Name:          key,
		}
		i++
	}
	createVaultInput := CreateVaultInput{
		AffiliationName: affiliation,
		Permissions:     vault.Permissions,
		VaultName:       vault.Name,
		Secrets:         secrets,
	}
	return createVaultInput
}

const addVaultPermissionsRequestString = `mutation addVaultPermissions($addVaultPermissionsInput: AddVaultPermissionsInput!){
  addVaultPermissions(input: $addVaultPermissionsInput)
  {
    hasAccess
    name
    permissions
  }
}`

// VaultPermissionsInput is input to the graphql addVaultPermissions and removeVAultPermissions interfaces
type VaultPermissionsInput struct {
	AffiliationName string   `json:"affiliationName"`
	Permissions     []string `json:"permissions"`
	VaultName       string   `json:"vaultName"`
}

// AddPermissions adds permissions to vault via gobo
func (api *APIClient) AddPermissions(vaultName string, permissions []string) error {
	addVaultPermissionsRequest := graphql.NewRequest(addVaultPermissionsRequestString)
	addVaultPermissionsInput := VaultPermissionsInput{
		AffiliationName: api.Affiliation,
		Permissions:     permissions,
		VaultName:       vaultName,
	}
	addVaultPermissionsRequest.Var("addVaultPermissionsInput", addVaultPermissionsInput)

	var addVaultPermissionsResponse VaultResponse
	if err := api.RunGraphQlMutation(addVaultPermissionsRequest, &addVaultPermissionsResponse); err != nil {
		return err
	}

	return nil
}

const removeVaultPermissionsRequestString = `mutation removeVaultPermissions($removeVaultPermissionsInput: RemoveVaultPermissionsInput!){
  removeVaultPermissions(input: $removeVaultPermissionsInput)
  {
    hasAccess
    name
    permissions
  }
}`

// RemovePermissions removes permissions from vault via gobo
func (api *APIClient) RemovePermissions(vaultName string, permissions []string) error {
	removeVaultPermissionsRequest := graphql.NewRequest(removeVaultPermissionsRequestString)
	removeVaultPermissionsInput := VaultPermissionsInput{
		AffiliationName: api.Affiliation,
		Permissions:     permissions,
		VaultName:       vaultName,
	}
	removeVaultPermissionsRequest.Var("removeVaultPermissionsInput", removeVaultPermissionsInput)

	var removeVaultPermissionsResponse VaultResponse
	if err := api.RunGraphQlMutation(removeVaultPermissionsRequest, &removeVaultPermissionsResponse); err != nil {
		return err
	}

	return nil
}

const deleteVaultRequestString = `mutation deleteVault($deleteVaultInput: DeleteVaultInput!){
  deleteVault(input: $deleteVaultInput)
  {
    affiliationName
    vaultName
  }
}`

// DeleteVaultInput is input to the graphql deleteVault interface
type DeleteVaultInput struct {
	AffiliationName string `json:"affiliationName"`
	VaultName       string `json:"vaultName"`
}

// DeleteVaultResponse is the response from the graphql "deleteVault"
type DeleteVaultResponse struct {
	AffiliationName string `json:"affiliationName"`
	VaultName       string `json:"vaultName"`
}

// DeleteVault deletes an aurora secret vault via API calls
func (api *APIClient) DeleteVault(vaultName string) error {
	deleteVaultRequest := graphql.NewRequest(deleteVaultRequestString)
	deleteVaultInput := DeleteVaultInput{
		AffiliationName: api.Affiliation,
		VaultName:       vaultName,
	}
	deleteVaultRequest.Var("deleteVaultInput", deleteVaultInput)

	var deleteVaultResponse DeleteVaultResponse
	if err := api.RunGraphQlMutation(deleteVaultRequest, &deleteVaultResponse); err != nil {
		return err
	}

	return nil
}

// GetSecret gets a secret by name
func (s Secrets) GetSecret(name string) (string, error) {
	secret, found := s[name]
	if !found {
		return "", errors.Errorf("Did not find secret %s", name)
	}
	data, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// AddSecret adds a secret
func (s Secrets) AddSecret(name, content string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	s[name] = encoded
}

// RemoveSecret deletes a secret
func (s Secrets) RemoveSecret(name string) {
	delete(s, name)
}
