package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save changed, new and deleted files for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := cmd.Flags().GetString("user")
		if _, err := auroraconfig.Save(user, config); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Save success")
		}
	},
}

func init() {
	RootCmd.AddCommand(saveCmd)

	viper.BindEnv("USER")
	saveCmd.Flags().StringP("user", "u", viper.GetString("USER"), "Save AuroraConfig as user")
}
