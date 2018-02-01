package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/editor"
	"github.com/skatteetaten/ao/pkg/fuzzy"
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
		return cmd.Usage()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	matches := fuzzy.FindMatches(search, fileNames, true)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched than one file. Search must be more specific.\n%v", matches)
	}

	fileName := matches[0]
	file, eTag, err := DefaultApiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	fileEditor := editor.NewEditor(func(modified string) ([]string, error) {
		file.Contents = modified
		res, err := DefaultApiClient.PutAuroraConfigFile(file, eTag)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res.GetAllErrors(), nil
		}
		return nil, nil
	})

	err = fileEditor.Edit(string(file.Contents), file.Name)
	if err != nil {
		return err
	}

	fmt.Println(fileName, "edited")
	return nil
}
