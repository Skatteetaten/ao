package client

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/skatteetaten/graphql"
)

const FoundNoSecretsForVault = "Found no secrets for vault"

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

type CreateVaultInput struct {
	AffiliationName string   `json:"affiliationName"`
	VaultName       string   `json:"vaultName"`
	Permissions     []string `json:"permissions"`
	Secrets         []Secret `json:"secrets"`
}

type CreateVaultResponse struct {
	CreateVault Vault `json:"createVault"`
}

func (api *APIClient) CreateVault(vault Vault) error {
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

func mapCreateVaultInput(vault Vault, affiliation string) CreateVaultInput {
	createVaultInput := CreateVaultInput{
		AffiliationName: affiliation,
		Permissions:     vault.Permissions,
		VaultName:       vault.Name,
		Secrets:         vault.Secrets,
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

type AddVaultSecretsInput struct {
	AffiliationName string   `json:"affiliationName"`
	Secrets         []Secret `json:"secrets"`
	VaultName       string   `json:"vaultName"`
}

// AddVaultSecretsResponse is core of response from graphql addVaultSecrets
type AddVaultSecretsResponse = Vault

const addVaultSecretsRequestString = `mutation addVaultSecrets ($addVaultSecretsInput: AddVaultSecretsInput!){
  addVaultSecrets(input: $addVaultSecretsInput)
  {
    name
    secrets {
		name
	}
  }
}`

// AddSecrets adds secrets to vault via gobo
func (api *APIClient) AddSecrets(vaultName string, secrets []Secret) error {
	addVaultSecretsRequest := graphql.NewRequest(addVaultSecretsRequestString)
	addVaultSecretsInput := AddVaultSecretsInput{
		AffiliationName: api.Affiliation,
		Secrets:         secrets,
		VaultName:       vaultName,
	}
	addVaultSecretsRequest.Var("addVaultSecretsInput", addVaultSecretsInput)

	var addVaultSecretsResponse AddVaultSecretsResponse
	if err := api.RunGraphQlMutation(addVaultSecretsRequest, &addVaultSecretsResponse); err != nil {
		return err
	}

	return nil
}

type RemoveVaultSecretsInput struct {
	AffiliationName string   `json:"affiliationName"`
	SecretNames     []string `json:"secretNames"`
	VaultName       string   `json:"vaultName"`
}

// RemoveVaultSecretsResponse is core of response from graphql removeVaultSecrets
type RemoveVaultSecretsResponse = Vault

const removeVaultSecretsRequestString = `mutation removeVaultSecrets ($removeVaultSecretsInput: RemoveVaultSecretsInput!){
  removeVaultSecrets(input: $removeVaultSecretsInput)
  {
    name
    secrets {
		name
	}
  }
}`

// RemoveSecrets removes secrets from vault via gobo
func (api *APIClient) RemoveSecrets(vaultName string, secretNames []string) error {
	removeVaultSecretsRequest := graphql.NewRequest(removeVaultSecretsRequestString)
	removeVaultSecretsInput := RemoveVaultSecretsInput{
		AffiliationName: api.Affiliation,
		SecretNames:     secretNames,
		VaultName:       vaultName,
	}
	removeVaultSecretsRequest.Var("removeVaultSecretsInput", removeVaultSecretsInput)

	var removeVaultSecretsResponse RemoveVaultSecretsResponse
	if err := api.RunGraphQlMutation(removeVaultSecretsRequest, &removeVaultSecretsResponse); err != nil {
		return err
	}

	return nil
}

type RenameVaultSecretInput struct {
	AffiliationName string `json:"affiliationName"`
	NewSecretName   string `json:"newSecretName"`
	SecretName      string `json:"secretName"`
	VaultName       string `json:"vaultName"`
}

// RenameVaultSecretResponse is core of response from graphql renameVaultSecret
type RenameVaultSecretResponse = Vault

const renameVaultSecretRequestString = `mutation renameVaultSecret ($renameVaultSecretInput: RenameVaultSecretInput!){
  renameVaultSecret(input: $renameVaultSecretInput)
  {
    name
    secrets {
		name
	}
  }
}`

// RenameSecret renames a secret in vault via gobo
func (api *APIClient) RenameSecret(vaultName, oldSecretName, newSecretName string) error {
	renameVaultSecretRequest := graphql.NewRequest(renameVaultSecretRequestString)
	renameVaultSecretInput := RenameVaultSecretInput{
		AffiliationName: api.Affiliation,
		NewSecretName:   newSecretName,
		SecretName:      oldSecretName,
		VaultName:       vaultName,
	}
	renameVaultSecretRequest.Var("renameVaultSecretInput", renameVaultSecretInput)

	var removeVaultSecretsResponse RenameVaultSecretResponse
	if err := api.RunGraphQlMutation(renameVaultSecretRequest, &removeVaultSecretsResponse); err != nil {
		return err
	}

	return nil
}

type UpdateVaultSecretInput struct {
	AffiliationName string `json:"affiliationName"`
	Base64Content   string `json:"base64Content"`
	SecretName      string `json:"secretName"`
	VaultName       string `json:"vaultName"`
}

// UpdateVaultSecretResponse is core of response from graphql updateVaultSecret
type UpdateVaultSecretResponse = Vault

const updateVaultSecretRequestString = `mutation updateVaultSecret ($updateVaultSecretInput: UpdateVaultSecretInput!){
  updateVaultSecret(input: $updateVaultSecretInput)
  {
    name
    secrets {
		name
	}
  }
}`

// UpdateSecret updates a secret in vault via gobo
func (api *APIClient) UpdateSecret(vaultName, secretName, modifiedContent string) error {
	updateVaultSecretRequest := graphql.NewRequest(updateVaultSecretRequestString)
	updateVaultSecretInput := UpdateVaultSecretInput{
		AffiliationName: api.Affiliation,
		Base64Content:   base64.StdEncoding.EncodeToString([]byte(modifiedContent)),
		SecretName:      secretName,
		VaultName:       vaultName,
	}
	updateVaultSecretRequest.Var("updateVaultSecretInput", updateVaultSecretInput)

	var updateVaultSecretsResponse UpdateVaultSecretResponse
	if err := api.RunGraphQlMutation(updateVaultSecretRequest, &updateVaultSecretsResponse); err != nil {
		return err
	}

	return nil
}
