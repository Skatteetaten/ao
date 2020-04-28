package cmd

import (
	"ao/pkg/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of all connected clusters",
	RunE:  Logout,
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}

// Logout performs the `logout` cli command
func Logout(cmd *cobra.Command, args []string) error {
	AO.Localhost = false
	AO.Affiliation = ""

	for _, c := range AO.Clusters {
		c.Token = ""
	}

	return config.WriteConfig(*AO, ConfigLocation)
}
