package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"os/user"
	_ "go/token"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone AuroraConfig (git repository) for current affiliation",
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

		if err := auroraconfig.Clone(affiliation, username, path); err != nil {
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
	cloneCmd.Flags().StringP("path", "p", "", "Clone repo to path")
	cloneCmd.Flags().StringP("user", "u", defaultUsername, "Clone repo as user")
}
