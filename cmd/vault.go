package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultAddGroup string
var vaultRemoveGroup string
var vaultAddUser string
var vaultRemoveUser string

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Create and perform operations on a vault",
	Long: `Usage:
vault create | edit | delete | permissions <vaultname>.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var vaultCrateCmd = &cobra.Command{
	Use:   "create <vaultname>",
	Short: "Create a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			vault.CreateVault(args[0], config)
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

var vaultEditCmd = &cobra.Command{
	Use:   "edit <vaultname>",
	Short: "Edit a vault",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.EditVault(args[0]); err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

var vaultPermissionsCmd = &cobra.Command{
	Use:   "permissions <vaultname>",
	Short: "Add or remove permissions on a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			if output, err := vault.Permissions(args[0], config, vaultAddGroup, vaultRemoveGroup, vaultAddUser, vaultRemoveUser); err == nil {
				fmt.Print(output)
			} else {
				fmt.Println(err.Error())
			}
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

func init() {
	RootCmd.AddCommand(vaultCmd)
	vaultCmd.AddCommand(vaultCrateCmd)
	vaultCmd.AddCommand(vaultEditCmd)

	vaultPermissionsCmd.Flags().StringVarP(&vaultAddGroup, "add-group", "", "", "Add a group permission to the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultRemoveGroup, "remove-group", "", "", "Remove a group permission from the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultAddUser, "add-user", "", "", "Add a user permission to the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultRemoveUser, "remove-user", "", "", "Remove a user permission from the vault")
	vaultCmd.AddCommand(vaultPermissionsCmd)
}
