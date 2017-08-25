package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"os/user"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {

		defaultUsername := ""
		if currentUser, err := user.Current(); err == nil {
			defaultUsername = currentUser.Username
		}

		if err := auroraconfig.Pull(defaultUsername); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Repository updated")
		}
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
