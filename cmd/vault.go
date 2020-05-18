package cmd

import (
	"fmt"

	"strings"

	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/editor"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

var (
	flagOnlyVaults bool

	errEmptyGroups            = errors.New("Cannot find groups in permissions")
	errNotValidSecretArgument = errors.New("not a valid argument, must be <vaultname/secret>")
)

var (
	vaultCmd = &cobra.Command{
		Use:         "vault",
		Short:       "Create and perform operations on a vault",
		Annotations: map[string]string{"type": "remote"},
	}

	vaultAddSecretCmd = &cobra.Command{
		Use:   "add-secret <vaultname> <[folder/]file>",
		Short: "Add a secret to an existing vault",
		RunE:  AddSecret,
	}

	vaultCreateCmd = &cobra.Command{
		Use:   "create <vaultname> <folder/file>",
		Short: "Create a new vault with secrets",
		RunE:  CreateVault,
	}

	vaultEditCmd = &cobra.Command{
		Use:   "edit-secret <vaultname/secret>",
		Short: "Edit an existing secret",
		RunE:  EditSecret,
	}

	vaultDeleteCmd = &cobra.Command{
		Use:   "delete <vaultname>",
		Short: "Delete a vault",
		RunE:  DeleteVault,
	}

	vaultDeleteSecretCmd = &cobra.Command{
		Use:   "delete-secret <vaultname/secret>",
		Short: "Delete a secret",
		RunE:  DeleteSecret,
	}

	vaultGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "List all vaults with permissions",
		Aliases: []string{"list"},
		RunE:    ListVaults,
	}

	vaultAddPermissionsCmd = &cobra.Command{
		Use:   "add-permissions <vaultname> <groups>",
		Short: "Add permissions on a vault",
		RunE:  VaultAddPermissions,
	}

	vaultRemovePermissionsCmd = &cobra.Command{
		Use:   "remove-permissions <vaultname> <groups>",
		Short: "Remove permissions on a vault",
		RunE:  VaultRemovePermissions,
	}

	vaultRenameCmd = &cobra.Command{
		Use:   "rename <vaultname> <new vaultname>",
		Short: "Rename a vault",
		RunE:  RenameVault,
	}
	vaultRenameSecretCmd = &cobra.Command{
		Use:   "rename-secret <vaultname/secretname> <new secretname>",
		Short: "Rename a secret",
		RunE:  RenameSecret,
	}
	vaultGetSecretCmd = &cobra.Command{
		Use:   "get-secret <vaultname/secretname>",
		Short: "Print the content of a secret to standard out",
		RunE:  GetSecret,
	}
)

func init() {
	RootCmd.AddCommand(vaultCmd)

	vaultCmd.AddCommand(vaultAddSecretCmd)
	vaultCmd.AddCommand(vaultAddPermissionsCmd)
	vaultCmd.AddCommand(vaultRemovePermissionsCmd)
	vaultCmd.AddCommand(vaultDeleteCmd)
	vaultCmd.AddCommand(vaultDeleteSecretCmd)
	vaultCmd.AddCommand(vaultGetCmd)
	vaultCmd.AddCommand(vaultCreateCmd)
	vaultCmd.AddCommand(vaultEditCmd)
	vaultCmd.AddCommand(vaultRenameCmd)
	vaultCmd.AddCommand(vaultRenameSecretCmd)
	vaultCmd.AddCommand(vaultGetSecretCmd)

	vaultGetCmd.Flags().BoolVarP(&flagAsList, "list", "", false, "print vault/secret as a list")
	vaultGetCmd.Flags().BoolVarP(&flagOnlyVaults, "only-vaults", "", false, "print vaults as a list")
}

// GetSecret is the entry point of the `vault get-secret` cli command
func GetSecret(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}
	split := strings.Split(args[0], "/")
	if len(split) != 2 {
		return errNotValidSecretArgument
	}
	vaultName, secretName := split[0], split[1]

	vault, err := DefaultAPIClient.GetVault(vaultName)
	if err != nil {
		return err
	}

	secret, err := vault.Secrets.GetSecret(secretName)
	if err != nil {
		return err
	}
	cmd.Printf("%s", secret)
	return nil
}

// AddSecret is the entry point of the `vault add-secret` cli command
func AddSecret(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	vault, err := DefaultAPIClient.GetVault(args[0])
	if err != nil {
		return err
	}

	err = collectSecrets(args[1], vault, false)
	if err != nil {
		return err
	}

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	cmd.Printf("New secrets has been added to vault %s\n", args[0])
	return nil
}

// RenameSecret is the entry point of the `vault rename-secret` cli command
func RenameSecret(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	split := strings.Split(args[0], "/")
	if len(split) != 2 {
		return errNotValidSecretArgument
	}

	newSecretName := args[1]
	vaultName, secretName := split[0], split[1]
	vault, err := DefaultAPIClient.GetVault(vaultName)
	if err != nil {
		return err
	}

	_, err = vault.Secrets.GetSecret(secretName)
	if err != nil {
		return err
	}

	_, ok := vault.Secrets[newSecretName]
	if ok {
		return errors.Errorf("Secret %s already exists\n", newSecretName)
	}

	vault.Secrets[newSecretName] = vault.Secrets[secretName]
	vault.Secrets.RemoveSecret(secretName)

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	cmd.Printf("Secret %s has been renamed to %s\n", secretName, newSecretName)
	return nil
}

// RenameVault is the entry point of the `vault rename` cli command
func RenameVault(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	vault, err := DefaultAPIClient.GetVault(args[1])
	if vault != nil {
		return errors.Errorf("Can't rename vault. %s already exists", args[1])
	}

	vault, err = DefaultAPIClient.GetVault(args[0])
	if err != nil {
		return err
	}

	vault.Name = args[1]

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	err = DefaultAPIClient.DeleteVault(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%s has been renamed to %s\n", args[0], args[1])
	return nil
}

// CreateVault is the entry point of the `vault create` cli command
func CreateVault(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	v, _ := DefaultAPIClient.GetVault(args[0])
	if v != nil {
		return errors.Errorf("vault %s already exists", args[0])
	}

	vault := client.NewAuroraSecretVault(args[0])

	err := collectSecrets(args[1], vault, true)
	if err != nil {
		return err
	}

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	fmt.Println("Vault", args[0], "created")
	return nil
}

// EditSecret is the entry point of the `vault edit-secret` cli command
func EditSecret(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	split := strings.Split(args[0], "/")
	if len(split) != 2 {
		return errNotValidSecretArgument
	}

	vaultName, secretName := split[0], split[1]
	contentToEdit, eTag, err := DefaultAPIClient.GetSecretFile(vaultName, secretName)
	if err != nil {
		return err
	}

	secretEditor := editor.NewEditor(func(modifiedContent string) error {
		return DefaultAPIClient.UpdateSecretFile(vaultName, secretName, eTag, []byte(modifiedContent))
	})

	err = secretEditor.Edit(contentToEdit, args[0])
	if err != nil {
		return err
	}

	cmd.Printf("Secret %s in vault %s edited\n", secretName, vaultName)
	return nil
}

// DeleteSecret is the entry point of the `vault delete-secret` cli command
func DeleteSecret(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	split := strings.Split(args[0], "/")
	if len(split) != 2 {
		return errNotValidSecretArgument
	}

	vaultName, secret := split[0], split[1]
	vault, err := DefaultAPIClient.GetVault(vaultName)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Do you want to delete secret %s?", args[0])
	shouldDelete := prompt.Confirm(message, false)
	if !shouldDelete {
		return nil
	}

	vault.Secrets.RemoveSecret(secret)

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	cmd.Printf("Secret %s deleted\n", args[0])
	return nil
}

// DeleteVault is the entry point of the `vault delete` cli command
func DeleteVault(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Usage()
	}

	message := fmt.Sprintf("Do you want to delete vault %s?", args[0])
	shouldDelete := prompt.Confirm(message, false)
	if !shouldDelete {
		return nil
	}

	err := DefaultAPIClient.DeleteVault(args[0])
	if err != nil {
		return err
	}

	cmd.Printf("Vault %s deleted\n", args[0])
	return nil
}

// ListVaults is the entry point of the `vault get` cli command
func ListVaults(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return cmd.Usage()
	}

	vaults, err := DefaultAPIClient.GetVaults()
	if err != nil {
		return err
	}

	var header string
	var rows []string
	if flagAsList {
		header = "VAULT/SECRET"
		for _, vault := range vaults {
			for secretName := range vault.Secrets {
				name := vault.Name + "/" + secretName
				rows = append(rows, name)
			}
		}
		sort.Strings(rows)
	} else if flagOnlyVaults {
		header = "VAULT"
		for _, vault := range vaults {
			rows = append(rows, vault.Name)
		}
		sort.Strings(rows)
	} else {
		header, rows = getVaultTable(vaults)
	}

	if len(rows) == 0 {
		return errors.New("No vaults available")
	}
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())

	return nil
}

// VaultAddPermissions is the entry point of the `vault add-permissions` cli command
func VaultAddPermissions(cmd *cobra.Command, args []string) error {
	return changeVaultPermissions(ADD, cmd, args)
}

// VaultRemovePermissions is the entry point of the `vault remove-permissions` cli command
func VaultRemovePermissions(cmd *cobra.Command, args []string) error {
	return changeVaultPermissions(DELETE, cmd, args)
}

type permissionAction uint64

// ADD and DELETE holds values for permission operations
const (
	ADD    permissionAction = 0
	DELETE permissionAction = 1
)

func changeVaultPermissions(action permissionAction, cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return cmd.Usage()
	}

	vault, err := DefaultAPIClient.GetVault(args[0])
	if err != nil {
		return err
	}

	vault.Permissions, err = handlePermissionAction(action, vault.Permissions, args[1:])
	if err != nil {
		return err
	}

	err = DefaultAPIClient.SaveVault(*vault)
	if err != nil {
		return err
	}

	cmd.Printf("Vault %s updated\n", args[0])
	return nil
}

func handlePermissionAction(action permissionAction, existingGroups, groups []string) ([]string, error) {

	modifiedGroups := existingGroups
	switch action {
	case ADD:
		{
			for _, group := range groups {
				for _, eg := range modifiedGroups {
					if eg == group {
						return nil, errors.Errorf("Group %s already exists", group)
					}
				}
				modifiedGroups = append(modifiedGroups, group)
			}
			return modifiedGroups, nil
		}
	case DELETE:
		{
			var found bool
			for _, group := range groups {
				found = false
				for i, g := range modifiedGroups {
					if g == group {
						modifiedGroups = append(modifiedGroups[:i], modifiedGroups[i+1:]...)
						found = true
					}
				}
				if !found {
					return nil, errors.Errorf("Did not find group %s", group)
				}
			}
			return modifiedGroups, nil
		}
	}

	return nil, nil
}

func collectSecrets(filePath string, vault *client.AuroraSecretVault, includePermissions bool) error {
	root, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	var files []os.FileInfo
	if root.IsDir() {
		files, err = ioutil.ReadDir(filePath)
		if err != nil {
			return err
		}
	} else {
		files = append(files, root)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		currentFilePath := filePath
		if root.IsDir() {
			currentFilePath = path.Join(filePath, f.Name())
		}

		if strings.Contains(f.Name(), "permission") && includePermissions {
			groups, err := readPermissionFile(currentFilePath)
			if err != nil {
				return err
			}
			vault.Permissions = groups
		} else {
			secret, err := readSecretFile(currentFilePath)
			if err != nil {
				return err
			}
			vault.Secrets.AddSecret(f.Name(), secret)
		}
	}

	return nil
}

func readSecretFile(fileName string) (string, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func readPermissionFile(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	permissions := struct {
		Groups []string `json:"groups"`
	}{}
	err = json.Unmarshal(data, &permissions)
	if err != nil {
		return nil, err
	}
	if permissions.Groups == nil {
		return nil, errEmptyGroups
	}

	return permissions.Groups, nil
}

func getVaultTable(vaults []*client.AuroraVaultInfo) (string, []string) {

	sort.Slice(vaults, func(i, j int) bool {
		return strings.Compare(vaults[i].Name, vaults[j].Name) < 1
	})

	var rows []string
	for _, vault := range vaults {
		name := vault.Name
		permissions := vault.Permissions

		if len(vault.Secrets) == 0 {
			line := fmt.Sprintf("%s\t%s\t%s\t%v", name, permissions, "", vault.HasAccess)
			rows = append(rows, line)
		} else {
			for secretName := range vault.Secrets {
				line := fmt.Sprintf("%s\t%s\t%s\t%v", name, permissions, secretName, vault.HasAccess)
				rows = append(rows, line)
				name = " "
			}
		}

	}

	header := "VAULT\tPERMISSIONS\tSECRET\tACCESS"
	return header, rows
}
