package cmd

import (
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

const addExample = `  Given the following AuroraConfig:
    - about.json
    - foobar.json
    - bar.json
    - foo/about.json
    - foo/bar.json
    - foo/foobar.json

  # adds test/about.json to AuroraConfig
  ao add test/about.json ./about.json

  # will throw an error because about.json already exists
  ao add about.json ./about.json

  # adds prod/about to AuroraConfig
  ao add prod/about ~/files/about.json`

var addCmd = &cobra.Command{
	Use:         "add <name> <file>",
	Short:       "Add a single file to the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     addExample,
	RunE:        Add,
}

func init() {
	RootCmd.AddCommand(addCmd)
}

func Add(cmd *cobra.Command, args []string) error {

	if len(args) != 2 {
		return cmd.Usage()
	}

	fileName, filePath := args[0], args[1]
	data, err := loadFile(filePath)
	if err != nil {
		return err
	}
	acf := &auroraconfig.File{
		Name:     fileName,
		Contents: string(data),
	}

	err = DefaultAPIClient.CreateAuroraConfigFile(acf)
	if err != nil {
		return err
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
