package cmd

import (
	"fmt"

	"os"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var affiliationSave bool

var affiliationCmd = &cobra.Command{
	Use:   "affiliation <Affiliation>",
	Short: "Change current affiliation",
	Long:  `Allows you to change the active affiliation.`,
	Run: func(cmd *cobra.Command, args []string) {
		var affiliation string
		var configLocation = viper.GetString("HOME") + "/.ao.json"

		if len(args) != 1 {
			fmt.Println("Please specify affiliation to change to")
			os.Exit(1)
		} else {
			affiliation = args[0]
		}
		openshift.Login(configLocation, userName, affiliation, apiCluster, persistentOptions.Localhost)
		if affiliationSave {
			user, _ := cmd.Flags().GetString("user")
			url := getGitUrl(config.GetAffiliation(), user)
			if _, err := auroraconfig.Save(url, config); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Save success")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(affiliationCmd)
	affiliationCmd.Flags().BoolVarP(&affiliationSave, "save", "", false,
		"Save updates on the affiliation to the local git repository")
	viper.BindEnv("USER")
	affiliationCmd.Flags().StringP("user", "u", viper.GetString("USER"), "Save AuroraConfig as user")
}
