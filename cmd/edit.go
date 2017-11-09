package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/command"
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

		filename, err := command.SelectFile(args[0], DefaultApiClient)
		if err != nil {
			fmt.Println(err)
			return
		}

		status, err := command.EditFile(filename, DefaultApiClient)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(status)
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
}
