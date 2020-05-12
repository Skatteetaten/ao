package cmd

import (
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

func AddGraphql(cmd *cobra.Command, args []string) error {

	if len(args) != 2 {
		return cmd.Usage()
	}

	fileName, filePath := args[0], args[1]
	data, err := loadFile(filePath)
	if err != nil {
		return err
	}

	createAuroraConfigFileRequestString := `mutation ($auroraConfigName: String!, $fileName: String!, $contents: String!){
  createAuroraConfigFile(input: {auroraConfigName: $auroraConfigName, fileName: $fileName, contents: $contents })
  {
    message
    success
  }
}`
	type ConfigFileValidationResponse struct {
		message string
		success string
	}

	createAuroraConfigFileRequest := graphql.NewRequest(createAuroraConfigFileRequestString)
	createAuroraConfigFileRequest.Var("auroraConfigName", DefaultApiClient.Affiliation)
	createAuroraConfigFileRequest.Var("fileName", fileName)
	createAuroraConfigFileRequest.Var("contents", string(data))

	var configFileValidationResponse ConfigFileValidationResponse
	if err := DefaultApiClient.RunGraphQlMutation(createAuroraConfigFileRequest, &configFileValidationResponse); err != nil {
		return err
	}

	// TODO: handle configFileValidationResponse

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
