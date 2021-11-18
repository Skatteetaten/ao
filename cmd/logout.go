package cmd

import (
	"github.com/skatteetaten/ao/pkg/session"
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
	AOSession.Localhost = false
	AOSession.AuroraConfig = ""
	AOSession.Tokens = map[string]string{}

	return session.WriteAOSession(*AOSession, SessionFileLocation)
}
