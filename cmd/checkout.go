package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	_ "go/token"
	"os/user"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Checkout AuroraConfig (git repository) for current affiliation",
	Run: func(cmd *cobra.Command, args []string) {
		affiliation := config.GetAffiliation()

		if affiliationFlag, _ := cmd.LocalFlags().GetString("affiliation"); len(affiliationFlag) > 0 {
			affiliation = affiliationFlag
		}

		username, _ := cmd.LocalFlags().GetString("user")
		path, _ := cmd.LocalFlags().GetString("path")

		if len(path) < 1 {
			path = fmt.Sprintf("./%s", affiliation)
		}

		if err := auroraconfig.Checkout(affiliation, username, path); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)

	defaultUsername := ""
	if currentUser, err := user.Current(); err == nil {
		defaultUsername = currentUser.Username
	}

	cloneCmd.Flags().StringP("affiliation", "a", "", "Affiliation to clone")
	cloneCmd.Flags().StringP("path", "p", "", "Checkout repo to path")
	cloneCmd.Flags().StringP("user", "u", defaultUsername, "Checkout repo as user")
}
