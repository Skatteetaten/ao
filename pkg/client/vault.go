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

	// AuroraSecretVault TODO: Deprecated.  Replace with Vault when Boober code is gone
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

const queryGetVaults = `
	query getVaults ($affiliation: String!) {
			 affiliations(name: $affiliation) {
    			edges {
      				node {
        				name
        				vaults {
          					name
          					permissions
							hasAccess
          					secrets {
            					name
          					}
        				}
      				}
				}
			 }
		}
`

func (api *APIClient) GetVaults() ([]Vault, error) {

	var respData AffiliationsResponse

	vars := map[string]interface{}{
		"affiliation": api.Affiliation,
	}

	if err := api.RunGraphQl(queryGetVaults, vars, &respData); err != nil {
		return nil, errors.Wrap(err, "Failed to get vaults.")
	}

	return respData.Vaults(api.Affiliation), nil
}

const queryGetSecretQuery = `
	query getVaults ($affiliation: String!, $vaultname: [String!]!, $secretname: [String!]!) {
		affiliations(name: $affiliation) {
			edges {
				node {
					name
					vaults(names: $vaultname){
						name
						secrets(names: $secretname){
							name
							base64Content
						}
					}
				}
			}
		}
	}
`

func (api *APIClient) GetSecret(vaultname, secretname string) (*Secret, error) {

	var respData AffiliationsResponse

	vars := map[string]interface{}{
		"affiliation": api.Affiliation,
		"vaultname":   []string{vaultname},
		"secretname":  []string{secretname},
	}

	if err := api.RunGraphQl(queryGetSecretQuery, vars, &respData); err != nil {
		return nil, errors.Wrap(err, "Failed to get secret")
	}

	secret := respData.Secret(api.Affiliation, vaultname, secretname)
	if secret == nil {
		return nil, errors.Errorf("Failed to find secret %s", secretname)
	}

	return secret, nil
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

type CreateVaultInput struct {
	AffiliationName string   `json:"affiliationName"`
	VaultName       string   `json:"vaultName"`
	Permissions     []string `json:"permissions"`
	Secrets         []Secret `json:"secrets"`
}

type CreateVaultResponse struct {
	CreateVault Vault `json:"createVault"`
}

func (api *APIClient) CreateVault(vault AuroraSecretVault) error {
	if len(vault.Permissions) == 0 {
		return errors.New("Aborted: Vault can not be created without permissions")
	}
	if len(vault.Secrets) == 0 {
		return errors.New(FoundNoSecretsForVault)
	}

	createVaultMutation := `
		mutation createVault($input: CreateVaultInput!) {
  			createVault(input: $input) {
				name
			}
		}
	`

	createVaultInput := mapCreateVaultInput(vault, api.Affiliation)

	createVaultRequest := graphql.NewRequest(createVaultMutation)
	createVaultRequest.Var("input", createVaultInput)

	var createVaultResponse CreateVaultResponse

	if err := api.RunGraphQlMutation(createVaultRequest, &createVaultResponse); err != nil {
		return err
	}

	return nil
}

func mapCreateVaultInput(vault AuroraSecretVault, affiliation string) CreateVaultInput {
	secrets := make([]Secret, len(vault.Secrets))
	i := 0
	for key, content := range vault.Secrets {
		secrets[i] = Secret{
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

type RenameVaultInput struct {
	AffiliationName string `json:"affiliationName"`
	VaultName       string `json:"vaultName"`
	NewVaultName    string `json:"newVaultName"`
}

type RenameVaultResponse struct {
	CreateVault Vault `json:"createVault"`
}

func (api *APIClient) RenameVault(oldVaultName, newVaultName string) error {

	renameVaultMutation := `
		mutation renameVault($renameVaultInput: RenameVaultInput!) {
  			renameVault(input: $renameVaultInput) {
				name
			}
		}
	`
	renameVaultInput := RenameVaultInput{
		AffiliationName: api.Affiliation,
		VaultName:       oldVaultName,
		NewVaultName:    newVaultName,
	}
	renameVaultRequest := graphql.NewRequest(renameVaultMutation)
	renameVaultRequest.Var("renameVaultInput", renameVaultInput)
	var createVaultResponse CreateVaultResponse

	if err := api.RunGraphQlMutation(renameVaultRequest, &createVaultResponse); err != nil {
		return err
	}

	return nil
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

type vaultPermissionsInput struct {
	AffiliationName string   `json:"affiliationName"`
	Permissions     []string `json:"permissions"`
	VaultName       string   `json:"vaultName"`
}

// AddVaultPermissionsInput is input to the graphql addVaultPermissions mutation
type AddVaultPermissionsInput = vaultPermissionsInput

// AddVaultPermissionsResponse is core of response from graphql addVaultPermissions
type AddVaultPermissionsResponse = Vault

const addVaultPermissionsRequestString = `mutation addVaultPermissions($addVaultPermissionsInput: AddVaultPermissionsInput!){
  addVaultPermissions(input: $addVaultPermissionsInput)
  {
    hasAccess
    name
    permissions
  }
}`

// AddPermissions adds permissions to vault via gobo
func (api *APIClient) AddPermissions(vaultName string, permissions []string) error {
	addVaultPermissionsRequest := graphql.NewRequest(addVaultPermissionsRequestString)
	addVaultPermissionsInput := AddVaultPermissionsInput{
		AffiliationName: api.Affiliation,
		Permissions:     permissions,
		VaultName:       vaultName,
	}
	addVaultPermissionsRequest.Var("addVaultPermissionsInput", addVaultPermissionsInput)

	var addVaultPermissionsResponse AddVaultPermissionsResponse
	if err := api.RunGraphQlMutation(addVaultPermissionsRequest, &addVaultPermissionsResponse); err != nil {
		return err
	}

	return nil
}

// RemoveVaultPermissionsInput is input to the graphql addVaultPermissions mutation
type RemoveVaultPermissionsInput = vaultPermissionsInput

// RemoveVaultPermissionsResponse is core of response from the graphql removeVaultPermissions
type RemoveVaultPermissionsResponse = Vault

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
	removeVaultPermissionsInput := RemoveVaultPermissionsInput{
		AffiliationName: api.Affiliation,
		Permissions:     permissions,
		VaultName:       vaultName,
	}
	removeVaultPermissionsRequest.Var("removeVaultPermissionsInput", removeVaultPermissionsInput)

	var removeVaultPermissionsResponse RemoveVaultPermissionsResponse
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
