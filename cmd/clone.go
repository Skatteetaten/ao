package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/skatteetaten/ao/pkg/configuration"
	"os/user"
	_ "go/token"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone AuroraConfig (git repository) for current affiliation",
	Run: func(cmd *cobra.Command, args []string) {

		var config configuration.ConfigurationClass
		affiliation := config.GetAffiliation()

		if affiliationFlag, _ := cmd.LocalFlags().GetString("affiliation"); len(affiliationFlag) > 0 {
			affiliation = affiliationFlag
		}

		if len(affiliation) < 1 {
			fmt.Println("No affiliation chosen, please login.")
			return
		}

		username, _ := cmd.LocalFlags().GetString("user")
		path, _ := cmd.LocalFlags().GetString("path")

		if len(path) < 1 {
			path = fmt.Sprintf("./%s", affiliation)
		}

		err := auroraconfig.Clone(affiliation, username, path)

		if err != nil {
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
