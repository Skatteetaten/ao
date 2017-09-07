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

var vaultFolder string

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Create and perform operations on a vault",
	Long: `Usage:
vault create | edit | delete | permissions | rename <vaultname> [<new vaultname>].`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var vaultCreateCmd = &cobra.Command{
	Use:   "create [<vaultname>] [-f <folder>]",
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

var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <vaultname>",
	Short: "Delete a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			if err := deletecmdObject.DeleteVault(args[0]); err == nil {
			} else {
				fmt.Println(err.Error())
			}
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

var vaultRenameCmd = &cobra.Command{
	Use:   "rename <vaultname> <new vaultname>",
	Short: "Rename a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 2 {

			if output, err := vault.Rename(args[0], args[1], config); err == nil {
				fmt.Println(output)
			} else {
				fmt.Println(err.Error())
			}
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

var vaultImportCmd = &cobra.Command{
	Use:   "import <catalog>",
	Short: "Imports the contents of a set of files into a set of vaults",
	Long: `Import works on a set of folders, each of which will become a separate vault.
Given the catalog structure:

vaultsfolder
  vault1
    secretfile1
    secretfile1
  vault2
    secretfile3

Then the command
	ao vault import vaultsfolder
will create 2 vaults: vault1 and vault2.  Vault1 will contain 2 secrets: secretfile1 and secretfile2.
Vault2 will contain 1 secret: secretfile3.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 1 {
			if output, err := vault.ImportVaults(args[0], config); err == nil {
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
	vaultCreateCmd.Flags().StringVarP(&vaultFolder, "folder", "f", "", "Creates a vault from a set of secret files")
	vaultCreateCmd.Flags().StringVarP(&vaultAddUser, "user", "u", "", "Adds a permission for the given user")
	vaultCreateCmd.Flags().StringVarP(&vaultAddGroup, "group", "g", "", "Adds a permission for the given group")

	vaultCmd.AddCommand(vaultCreateCmd)
	vaultCmd.AddCommand(vaultEditCmd)
	vaultCmd.AddCommand(vaultRenameCmd)

	vaultPermissionsCmd.Flags().StringVarP(&vaultAddGroup, "add-group", "", "", "Add a group permission to the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultRemoveGroup, "remove-group", "", "", "Remove a group permission from the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultAddUser, "add-user", "", "", "Add a user permission to the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultRemoveUser, "remove-user", "", "", "Remove a user permission from the vault")
	vaultCmd.AddCommand(vaultPermissionsCmd)
	vaultCmd.AddCommand(vaultDeleteCmd)

	vaultCmd.AddCommand(vaultImportCmd)
}
