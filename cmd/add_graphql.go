package cmd

import (
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

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

// AuroraConfigFileValidationResponse is core of response from the graphql "createAuroraConfigFile"
type AuroraConfigFileValidationResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// CreateAuroraConfigFileResponse is response from the named graphql mutation "createAuroraConfigFile"
type CreateAuroraConfigFileResponse struct {
	CreateAuroraConfigFile AuroraConfigFileValidationResponse `json:"createAuroraConfigFile"`
}

func AddGraphql(cmd *cobra.Command, args []string) error {

	if len(args) != 2 {
		return cmd.Usage()
	}

	fileName, filePath := args[0], args[1]
	data, err := loadFile(filePath)
	if err != nil {
		return err
	}

	createAuroraConfigFileRequest := graphql.NewRequest(createAuroraConfigFileRequestString)
	newAuroraConfigFileInput := NewAuroraConfigFileInput{
		AuroraConfigName: DefaultAPIClient.Affiliation,
		FileName:         fileName,
		Contents:         string(data),
	}
	createAuroraConfigFileRequest.Var("newAuroraConfigFileInput", newAuroraConfigFileInput)

	var createAuroraConfigFileResponse CreateAuroraConfigFileResponse
	if err := DefaultAPIClient.RunGraphQlMutation(createAuroraConfigFileRequest, &createAuroraConfigFileResponse); err != nil {
		return err
	}

	if !createAuroraConfigFileResponse.CreateAuroraConfigFile.Success {
		return errors.Errorf("Remote error: %s\n", createAuroraConfigFileResponse.CreateAuroraConfigFile.Message)
	}

	cmd.Printf("%s has been added\n", fileName)

	return nil
}

func loadFile(filePath string) ([]byte, error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if file.IsDir() {
		return nil, errors.New("Only files are supported")
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
