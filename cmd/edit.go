package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [env/]file",
	Short: "Edit a single file in the AuroraConfig repository, or a secret in a vault",
	Long: `Edit a single file in the AuroraConfig repository, or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`,
	Annotations: map[string]string{
		"type": "file",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		fileNames, err := DefaultApiClient.GetFileNames()
		if err != nil {
			fmt.Println(err)
			return
		}

		options, err := fuzzy.SearchForFile(args[0], fileNames)
		if err != nil {
			fmt.Println(err)
			return
		}

		filename := ""
		if len(options) > 1 {
			filename = prompt.SelectFile(options)
		} else if len(options) == 1 {
			filename = options[0]
		}

		if filename == "" {
			fmt.Println("No file to edit")
		}

		status, err := command.EditFile(filename, *DefaultApiClient)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(status)
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
