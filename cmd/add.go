package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
  ao add prod/about ~/files/about.json

  '.json' can be omitted from name, will be added if missing`

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

	fileName := args[0]
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	filePath := args[1]
	file, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if file.IsDir() {
		return errors.New("Only files are supported")
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if !json.Valid(data) {
		return errors.Errorf("%s contains illegal json format\n", filePath)
	}

	ac, err := DefaultApiClient.GetAuroraConfig()
	if err != nil {
		return err
	}

	if _, ok := ac.Files[fileName]; ok {
		return errors.Errorf("File %s already exists\n", fileName)
	}

	ac.Files[fileName] = data

	res, err := DefaultApiClient.SaveAuroraConfig(ac)
	if err != nil {
		return err
	}
	if res != nil {
		return errors.New(res.String())
	}

	cmd.Printf("%s has been added\n", fileName)

	return nil
}
