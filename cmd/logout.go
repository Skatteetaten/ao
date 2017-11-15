package cmd

import (
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of all connected clusters",
	Long:  `Removes the tokens stored for each cluster in the configuration file.`,
	RunE:  Logout,
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}

func Logout(cmd *cobra.Command, args []string) error {
	AO.Affiliation = ""
	for _, c := range AO.Clusters {
		c.Token = ""
	}

	AO.Localhost = false

	return config.WriteConfig(*AO, ConfigLocation)
}
