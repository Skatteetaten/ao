package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/editor"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

const editFileLong = `Edit a single file in the AuroraConfig repository, or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`

var editCmd = &cobra.Command{
	Use:         "edit [env/]file",
	Short:       "Edit a single file in the AuroraConfig repository, or a secret in a vault",
	Long:        editFileLong,
	Annotations: map[string]string{"type": "file"},
	RunE:        EditFile,
}

func init() {
	RootCmd.AddCommand(editCmd)
}

func EditFile(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Usage()
		return nil
	}

	fileName, err := SelectFile(args[0], DefaultApiClient)
	if err != nil {
		return err
	}

	file, err := DefaultApiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	err = editor.Edit(string(file.Contents), file.Name, true, func(modified string) ([]string, error) {
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

	if err != nil {
		return err
	}

	fmt.Println(fileName, "edited")
	return nil
}

func SelectFile(search string, api *client.ApiClient) (string, error) {
	var fileName string
	fileNames, err := api.GetFileNames()
	if err != nil {
		return fileName, err
	}

	options, err := fuzzy.SearchForFile(search, fileNames)
	if err != nil {
		return fileName, err
	}

	if len(options) > 1 {
		message := fmt.Sprintf("Matched %d files. Which file do you want?", len(options))
		fileName = prompt.Select(message, options)
	} else if len(options) == 1 {
		fileName = options[0]
	}

	if fileName == "" {
		return fileName, errors.New("No files")
	}

	return fileName, nil
}
