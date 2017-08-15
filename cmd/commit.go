package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"os/user"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changed, new and deleted files for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {

		defaultUsername := ""
		if currentUser, err := user.Current(); err == nil {
			defaultUsername = currentUser.Username
		}

		if err := auroraconfig.Commit(defaultUsername, &persistentOptions); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Commit success")
		}
	},
}

func init() {
	RootCmd.AddCommand(commitCmd)
}
