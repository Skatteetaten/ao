package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/vault"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var createVaultCmd = &cobra.Command{
	Use:   "vault [<vaultname>] [-f <folder>]",
	Short: "Creates a vault and optionally imports the contents of a set of secretfiles into a vault",
	Long: `Create <vaultname> will create an empty vault.
If the --folder / -f flag is given, ao will read all the files in <folder>, and each file will become a secret.
The secret will be named the same as the file.
If no vaultname is given, the vault will be named the same as the <folder>.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			vaultname := ""
			if len(args) == 1 {
				vaultname = args[0]
			}
			vault.CreateVault(vaultname, config, vaultFolder, vaultAddUser, vaultAddGroup)
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createVaultCmd)
	createVaultCmd.Flags().StringVarP(&vaultFolder, "folder", "f", "", "Creates a vault from a set of secret files")
	createVaultCmd.Flags().StringVarP(&vaultAddUser, "user", "u", "", "Adds a permission for the given user")
	createVaultCmd.Flags().StringVarP(&vaultAddGroup, "group", "g", "", "Adds a permission for the given group")
	createCmd.Hidden = true
}
