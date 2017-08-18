package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/skatteetaten/ao/pkg/vault"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a resource",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var createVaultCmd = &cobra.Command{
	Use:   "vault <vaultname>",
	Short: "Create a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			vault.CreateVault(args[0], config)
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createVaultCmd)
}
