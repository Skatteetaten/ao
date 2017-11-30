package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/skatteetaten/ao/cmd/common"
	"github.com/skatteetaten/ao/pkg/editor"
	"github.com/spf13/cobra"
)

const editLong = `Edit a single file in the current AuroraConfig.`

const exampleEdit = `  Given the following AuroraConfig:
    - about.json
    - foobar.json
    - bar.json
    - foo/about.json
    - foo/bar.json
    - foo/foobar.json

  # Exact matching: will open foo/bar.json in editor
  ao edit foo/bar

  # Fuzzy matching: will open foo/foobar.json in editor
  ao edit fofoba
`

var editCmd = &cobra.Command{
	Use:         "edit [env/]file",
	Short:       "Edit a single file in the AuroraConfig repository",
	Long:        editLong,
	Annotations: map[string]string{"type": "remote"},
	Example:     exampleEdit,
	RunE:        EditFile,
}

func init() {
	RootCmd.AddCommand(editCmd)
}

func EditFile(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Choose a file to edit")
	fileName, err := common.SelectOne(message, args, fileNames, true)
	if err != nil {
		return err
	}

	file, err := DefaultApiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	fileEditor := editor.NewEditor(func(modified string) ([]string, error) {
		file.Contents = json.RawMessage(modified)
		res, err := DefaultApiClient.PutAuroraConfigFile(file)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res.GetAllErrors(), nil
		}
		return nil, nil
	})

	err = fileEditor.Edit(string(file.Contents), file.Name, true)
	if err != nil {
		return err
	}

	fmt.Println(fileName, "edited")
	return nil
}
