package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/skatteetaten/ao/pkg/vault"
)

var createCmd = &cobra.Command{
	Use:   "create vault <vaultname> | secret <vaultname> <secretname>",
	Short: "Creates a vault or a secret in a vault",
	Long:  `Creates a vault or a secret in a vault.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UseLine())
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
