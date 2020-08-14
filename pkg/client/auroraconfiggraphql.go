package client

import (
	"encoding/json"
	"github.com/machinebox/graphql"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

// AuroraConfigClientGraphql is a an internal client facade for external aurora configuration API calls using graphql
type AuroraConfigClientGraphql interface {
	CreateAuroraConfigFile(fileName string, data []byte) (CreateAuroraConfigFileResponse, error)
	UpdateAuroraConfigFile(file *auroraconfig.File, eTag string) (UpdateAuroraConfigFileResponse, error)
}

// AuroraConfigFileValidationResponse is core of response from the graphql "createAuroraConfigFile" and "updateAuroraConfigFile"
type AuroraConfigFileValidationResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

const createAuroraConfigFileRequestString = `mutation createAuroraConfigFile($newAuroraConfigFileInput: NewAuroraConfigFileInput!){
  createAuroraConfigFile(input: $newAuroraConfigFileInput)
  {
    message
    success
  }
}`

// NewAuroraConfigFileInput is input to the graphql createAuroraConfigFile interface
type NewAuroraConfigFileInput struct {
	AuroraConfigName string `json:"auroraConfigName"`
	FileName         string `json:"fileName"`
	Contents         string `json:"contents"`
}

// CreateAuroraConfigFileResponse is response from the named graphql mutation "createAuroraConfigFile"
type CreateAuroraConfigFileResponse struct {
	CreateAuroraConfigFile AuroraConfigFileValidationResponse `json:"createAuroraConfigFile"`
}

// CreateAuroraConfigFile creates an Aurora config file via API call (graphql)
func (api *APIClient) CreateAuroraConfigFile(fileName string, data []byte) (*CreateAuroraConfigFileResponse, error) {
	createAuroraConfigFileRequest := graphql.NewRequest(createAuroraConfigFileRequestString)
	newAuroraConfigFileInput := NewAuroraConfigFileInput{
		AuroraConfigName: api.Affiliation,
		FileName:         fileName,
		Contents:         string(data),
	}
	createAuroraConfigFileRequest.Var("newAuroraConfigFileInput", newAuroraConfigFileInput)

	var createAuroraConfigFileResponse CreateAuroraConfigFileResponse
	if err := api.RunGraphQlMutation(createAuroraConfigFileRequest, &createAuroraConfigFileResponse); err != nil {
		return nil, err
	}
	return &createAuroraConfigFileResponse, nil
}

const updateAuroraConfigFileRequestString = `mutation updateAuroraConfigFile($updateAuroraConfigFileInput: UpdateAuroraConfigFileInput!){
  updateAuroraConfigFile(input: $updateAuroraConfigFileInput)
  {
    message
    success
  }
}`

// UpdateAuroraConfigFileInput is input to the graphql updateAuroraConfigFile interface
type UpdateAuroraConfigFileInput struct {
	AuroraConfigName string `json:"auroraConfigName"`
	FileName         string `json:"fileName"`
	Contents         string `json:"contents"`
	ExistingHash     string `json:"existingHash"`
}

// UpdateAuroraConfigFileResponse is response from the named graphql mutation "updateAuroraConfigFile"
type UpdateAuroraConfigFileResponse struct {
	UpdateAuroraConfigFile AuroraConfigFileValidationResponse `json:"updateAuroraConfigFile"`
}

// UpdateAuroraConfigFile updates an Aurora config file via API call (graphql)
func (api *APIClient) UpdateAuroraConfigFile(file *auroraconfig.File, eTag string) (UpdateAuroraConfigFileResponse, error) {
	logrus.Debugf("UpdateAuroraConfigFile: ETag: %s", eTag)
	updateAuroraConfigFileRequest := graphql.NewRequest(updateAuroraConfigFileRequestString)

	if err := validateFileContentIsJSON(file); err != nil {
		return UpdateAuroraConfigFileResponse{}, err
	}

	updateAuroraConfigFileInput := UpdateAuroraConfigFileInput{
		AuroraConfigName: api.Affiliation,
		FileName:         file.Name,
		Contents:         file.Contents,
		ExistingHash:     "",
	}
	if eTag != "" {
		updateAuroraConfigFileInput.ExistingHash = eTag
	}

	updateAuroraConfigFileRequest.Var("updateAuroraConfigFileInput", updateAuroraConfigFileInput)

	var updateAuroraConfigFileResponse UpdateAuroraConfigFileResponse
	if err := api.RunGraphQlMutation(updateAuroraConfigFileRequest, &updateAuroraConfigFileResponse); err != nil {
		return UpdateAuroraConfigFileResponse{}, err
	}
	return updateAuroraConfigFileResponse, nil
}

func validateFileContentIsJSON(file *auroraconfig.File) error {
	payload := auroraConfigFilePayload{
		Content: string(file.Contents),
	}
	_, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return nil
}
