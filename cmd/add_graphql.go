package cmd

import (
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

	createAuroraConfigFileResponse, err := DefaultAPIClient.CreateAuroraConfigFile(fileName, data)
	if err != nil {
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
