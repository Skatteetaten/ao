package cmd

import (
	"fmt"

	"strings"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/editor"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path"
	"sort"
)

const createVaultLong = `Create <vaultname> will create an empty vault.
If the --folder / -f flag is given, ao will read all the files in <folder>, and each file will become a secret.
The secret will be named the same as the file.
If no vaultname is given, the vault will be named the same as the <folder>.`

const editVaultLong = `This command will edit the content of the given vault.
The editor will present a JSON view of the vault.
The secrets will be presented as Base64 encoded strings.
If secret-name is given, the editor will present the decoded secret string for editing.`

const importVaultLong = `Import works on a set of folders, each of which will become a separate vault.
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
Vault2 will contain 1 secret: secretfile3.`

const listVaultLong = `If no argument is given, the command will list the vaults in the current affiliation, along with the
numer of secrets in the vault.
If a vaultname is specified, the command will list the secrets in the given vault.
To access a secret, use the get secret command.`

var (
	vaultAddGroup    string
	vaultRemoveGroup string
	vaultFolder      string
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Create and perform operations on a vault",
}

var vaultCreateCmd = &cobra.Command{
	Use:   "create [<vaultname>] [-f <folder>]",
	Short: "Creates a vault and optionally imports the contents of a set of secretfiles into a vault",
	Long:  createVaultLong,
	RunE:  CreateVault,
}

var vaultEditCmd = &cobra.Command{
	Use:   "edit <vaultname> | <vaultname>/<secretname> | <vaultname> <secretname>",
	Short: "Edit a vault or a secret",
	Long:  editVaultLong,
	RunE:  EditVault,
}

var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <vaultname>",
	Short: "Delete a vault",
	RunE:  DeleteVault,
}

var vaultImportCmd = &cobra.Command{
	Use:   "import <catalog>",
	Short: "Imports the contents of a set of files into a set of vaults",
	Long:  importVaultLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Not implemented")
	},
}

var vaultListCmd = &cobra.Command{
	Use:     "list [vaultname]",
	Short:   "list all vaults",
	Aliases: []string{"vaults"},
	Long:    listVaultLong,
	RunE:    ListVaults,
}

var vaultPermissionsCmd = &cobra.Command{
	Use:   "permissions <vaultname>",
	Short: "Add or remove permissions on a vault",
	RunE:  VaultPermissions,
}

var vaultRenameCmd = &cobra.Command{
	Use:   "rename <vaultname> <new vaultname>",
	Short: "Rename a vault",
	RunE:  RenameVault,
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

func RenameVault(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	vault, err := DefaultApiClient.GetVault(args[1])
	if vault != nil {
		return errors.Errorf("Can't rename vault. %s already exists", args[1])
	}

	vault, err = DefaultApiClient.GetVault(args[0])
	if err != nil {
		return err
	}

	vault.Name = args[1]

	err = DefaultApiClient.SaveVault(*vault, false)
	if err != nil {
		return err
	}

	err = DefaultApiClient.DeleteVault(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%s has been renamed to %s\n", args[0], args[1])
	return nil
}

func CreateVault(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	// TODO: Check if vault exists
	vault := client.NewAuroraSecretVault(args[0])

	if vaultFolder != "" {
		err := collectSecrets(vaultFolder, vault)
		if err != nil {
			return err
		}
	}

	if vaultRemoveGroup != "" {
		err := vault.Permissions.DeleteGroup(vaultRemoveGroup)
		if err != nil {
			return err
		}
	}

	if vaultAddGroup != "" {
		err := vault.Permissions.AddGroup(vaultAddGroup)
		if err != nil {
			return err
		}
	}

	err := DefaultApiClient.SaveVault(*vault, false)
	if err != nil {
		return err
	}

	fmt.Println("Vault", args[0], "created")
	return nil
}

func EditVault(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return cmd.Usage()
	}

	vaultName := args[0]
	var secretName string
	if len(args) == 2 {
		secretName = args[1]
	}

	var contentToEdit string
	vault, err := DefaultApiClient.GetVault(vaultName)
	if err != nil {
		return err
	}

	if secretName == "" {
		data, err := json.Marshal(vault)
		if err != nil {
			return err
		}
		contentToEdit = string(data)
	} else {
		contentToEdit, err = vault.Secrets.GetSecret(secretName)
		if err != nil {
			return err
		}
	}

	name := vaultName + " " + secretName
	err = editor.Edit(contentToEdit, name, false, func(modifiedContent string) ([]string, error) {
		if secretName == "" {
			err := json.Unmarshal([]byte(modifiedContent), &vault)
			if err != nil {
				return nil, err
			}
		} else {
			vault.Secrets.AddSecret(secretName, modifiedContent)
		}
		err := DefaultApiClient.SaveVault(*vault, true)
		if err != nil {
			return []string{err.Error()}, nil
		}
		return nil, nil
	})

	if err != nil {
		return err
	}

	fmt.Println("Secret saved")

	return nil
}

func DeleteVault(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Usage()
	}

	message := fmt.Sprintf("Do you want to delete vault %s?", args[0])
	shouldDelete := prompt.Confirm(message)
	if !shouldDelete {
		return nil
	}

	err := DefaultApiClient.DeleteVault(args[0])
	if err != nil {
		return err
	}

	fmt.Println("Delete success")
	return nil
}

func ListVaults(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return cmd.Usage()
	}

	vaults, err := DefaultApiClient.GetVaults()
	if err != nil {
		return err
	}

	table := getVaultTable(vaults)
	if len(table) == 1 {
		return errors.New("No vaults available")
	}
	DefaultTablePrinter(table)

	return nil
}

func VaultPermissions(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Usage()
	}

	vault, err := DefaultApiClient.GetVault(args[0])
	if err != nil {
		return err
	}

	if vaultRemoveGroup != "" {
		err = vault.Permissions.DeleteGroup(vaultRemoveGroup)
		if err != nil {
			return err
		}
	}

	if vaultAddGroup != "" {
		err = vault.Permissions.AddGroup(vaultAddGroup)
		if err != nil {
			return err
		}
	}

	err = DefaultApiClient.SaveVault(*vault, true)
	if err != nil {
		return err
	}

	fmt.Println("Vault saved")
	return nil
}

func collectSecrets(folder string, vault *client.AuroraSecretVault) error {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(path.Join(folder, f.Name()))
		if err != nil {
			return err
		}
		if !strings.Contains(f.Name(), "permission") {
			vault.Secrets.AddSecret(f.Name(), string(data))
			continue
		}

		permissions := struct {
			Groups []string `json:"groups"`
		}{}

		err = json.Unmarshal(data, &permissions)
		if err != nil {
			return err
		}
		if permissions.Groups == nil {
			return errors.New("Cannot find groups in permissions")
		}
		vault.Permissions["groups"] = permissions.Groups
	}

	return nil
}

func getVaultTable(vaults []*client.AuroraVaultInfo) []string {
	table := []string{"VAULT\tPERMISSIONS\tSECRET\tACCESS"}

	sort.Slice(vaults, func(i, j int) bool {
		return strings.Compare(vaults[i].Name, vaults[j].Name) < 1
	})

	for _, vault := range vaults {
		name := vault.Name
		permissions := vault.Permissions.GetGroups()

		for _, s := range vault.Secrets {
			line := fmt.Sprintf("%s\t%s\t%s\t%v\t", name, permissions, s, vault.Admin)
			table = append(table, line)
			name = " "
		}
	}

	return table
}
