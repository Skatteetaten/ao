package cmd

import (
	"fmt"

	"strings"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/skatteetaten/ao/pkg/editcmd"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/skatteetaten/ao/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultAddGroup string
var vaultRemoveGroup string
var vaultAddUser string

var vaultFolder string

var editCmdObject = &editcmd.EditcmdClass{
	Configuration: oldConfig,
}

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
			vault.CreateVault(vaultname, oldConfig, vaultFolder, vaultAddUser, vaultAddGroup)
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

var vaultEditCmd = &cobra.Command{
	Use:   "edit <vaultname> | <vaultname>/<secretname> | <vaultname> <secretname>",
	Short: "Edit a vault or a secret",
	Long: `This command will edit the content of the given vault.
The editor will present a JSON view of the vault.
The secrets will be presented as Base64 encoded strings.
If secret-name is given, the editor will present the decoded secret string for editing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var vaultname string
		var secretname string
		var output string
		var err error
		if len(args) == 1 {
			if strings.Contains(args[0], "/") {
				parts := strings.Split(args[0], "/")
				vaultname = parts[0]
				secretname = parts[1]
			} else {
				vaultname = args[0]
			}
		} else if len(args) == 2 {
			vaultname = args[0]
			secretname = args[1]
		} else {
			cmd.Usage()
			return
		}

		if secretname != "" {
			output, err = editCmdObject.EditSecret(vaultname, secretname)
		} else {
			output, err = editCmdObject.EditVault(vaultname)
		}
		if err == nil {
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
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		vault, err := DefaultApiClient.GetVault(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		if vaultRemoveGroup != "" {
			err = vault.Permissions.DeleteGroup(vaultRemoveGroup)
			if err != nil {
				fmt.Println(err)
			}
		}

		if vaultAddGroup != "" {
			err = vault.Permissions.AddGroup(vaultAddGroup)
			if err != nil {
				fmt.Println(err)
			}
		}

		// TODO: Save vault
	},
}

var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <vaultname>",
	Short: "Delete a vault",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		message := fmt.Sprintf("Do you want to delete vault %s?", args[0])
		shouldDelete := prompt.Confirm(message)
		if !shouldDelete {
			return
		}

		err := DefaultApiClient.DeleteVault(args[0])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Delete success")
		}
	},
}

var vaultRenameCmd = &cobra.Command{
	Use:   "rename <vaultname> <new vaultname>",
	Short: "Rename a vault",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 2 {

			if output, err := vault.Rename(args[0], args[1], oldConfig); err == nil {
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
			if output, err := vault.ImportVaults(args[0], oldConfig); err == nil {
				fmt.Print(output)
			} else {
				fmt.Println(err.Error())
			}
		} else {
			fmt.Println(cmd.UseLine())
		}
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list [vaultname]",
	Short: "list all vaults",
	Long: `If no argument is given, the command will list the vaults in the current affiliation, along with the
numer of secrets in the vault.
If a vaultname is specified, the command will list the secrets in the given vault.
To access a secret, use the get secret command.`,
	Aliases: []string{"vaults"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			cmd.Usage()
			return
		}

		var vaults []*client.AuroraSecretVault
		if len(args) > 0 {
			vault, err := DefaultApiClient.GetVault(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
			vaults = []*client.AuroraSecretVault{vault}
		} else {
			var err error
			vaults, err = DefaultApiClient.GetVaults()
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		table := command.GetVaultTable(vaults)
		if len(table) == 1 {
			fmt.Println("No vaults available")
			return
		}
		command.DefaultTablePrinter(table)
	},
}

func init() {
	RootCmd.AddCommand(vaultCmd)

	vaultCmd.AddCommand(vaultPermissionsCmd)
	vaultCmd.AddCommand(vaultDeleteCmd)
	vaultCmd.AddCommand(vaultImportCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultCreateCmd)
	vaultCmd.AddCommand(vaultEditCmd)
	vaultCmd.AddCommand(vaultRenameCmd)

	vaultCreateCmd.Flags().StringVarP(&vaultFolder, "folder", "f", "", "Creates a vault from a set of secret files")
	vaultPermissionsCmd.Flags().StringVarP(&vaultAddGroup, "add-group", "", "", "Add a group permission to the vault")
	vaultPermissionsCmd.Flags().StringVarP(&vaultRemoveGroup, "remove-group", "", "", "Remove a group permission from the vault")
}
