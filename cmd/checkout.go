package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go/token"
	"os/user"
	"os"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Checkout AuroraConfig (git repository) for current affiliation",
	Run: func(cmd *cobra.Command, args []string) {
		affiliation := config.GetAffiliation()

		if affiliationFlag, _ := cmd.LocalFlags().GetString("affiliation"); len(affiliationFlag) > 0 {
			affiliation = affiliationFlag
		}

		username, _ := cmd.LocalFlags().GetString("user")
		path, _ := cmd.LocalFlags().GetString("path")

		if len(path) < 1 {
			wd, _ := os.Getwd()
			path = fmt.Sprintf("%s/%s", wd, affiliation)
		}

		if err := auroraconfig.Checkout(affiliation, username, path); err != nil {
			fmt.Println(err)
			return
		}

		var configLocation = viper.GetString("HOME") + "/.ao.json"
		config, _ := openshift.LoadOrInitiateConfigFile(configLocation, false)

		if err := config.AddCheckoutPath(affiliation, path, configLocation); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(checkoutCmd)

	defaultUsername := ""
	if currentUser, err := user.Current(); err == nil {
		defaultUsername = currentUser.Username
	}

	checkoutCmd.Flags().StringP("affiliation", "a", "", "Affiliation to clone")
	checkoutCmd.Flags().StringP("path", "p", "", "Checkout repo to path")
	checkoutCmd.Flags().StringP("user", "u", defaultUsername, "Checkout repo as user")
}
