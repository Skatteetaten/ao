package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
	"os"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save changed, new and deleted files for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := cmd.Flags().GetString("user")
		url := getGitUrl(ao.Affiliation, user)

		if _, err := versioncontrol.Save(url, DefaultApiClient); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Save success")
		}
	},
}

func init() {
	RootCmd.AddCommand(saveCmd)

	user, _ := os.LookupEnv("USER")
	saveCmd.Flags().StringP("user", "u", user, "Save AuroraConfig as user")
}
