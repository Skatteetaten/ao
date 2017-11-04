package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of all connected clusters",
	Long:  `Removes the tokens stored for each cluster in the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ao.Logout(configLocation)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
