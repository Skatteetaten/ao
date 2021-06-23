package client

import "github.com/skatteetaten/graphql"

// TODO: Rewrite everything for Redeploy (not Vault)

type RedeployInput struct {
	AffiliationName string `json:"affiliationName"`
	VaultName       string `json:"vaultName"`
	NewVaultName    string `json:"newVaultName"`
}

type RedeployResponse struct {
	CreateVault Vault `json:"createVault"`
}

func (api *APIClient) Redeploy(oldVaultName, newVaultName string) error {

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