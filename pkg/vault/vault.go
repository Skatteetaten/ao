package vault

import (
	"errors"

	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"strings"

	"encoding/json"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/printutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

/*
type Vault struct {
	vaults []serverapi.Vault
}
*/

func CreateVault(vaultname string, config *configuration.ConfigurationClass, folderName string, addUser string, addGroup string) (output string, err error) {
	var vault serverapi.Vault

	if folderName == "" {
		vault.Name = vaultname
		vault.Secrets = make(map[string]string)
	} else {
		vault, err = secretsFolder2Vault(folderName)
		if err != nil {
			return "", err
		}
		// Override vaultname if given
		if vaultname != "" {
			vault.Name = vaultname
		}
	}
	// Add permissions if specified
	if addUser != "" {
		vault.Permissions.Users = append(vault.Permissions.Users, addUser)
	}
	if addGroup != "" {
		vault.Permissions.Groups = append(vault.Permissions.Groups, addGroup)
	}

	exists, err := vaultExists(vault.Name, config)
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New("Error: Vault " + vault.Name + " exists")
	}

	message, err := auroraconfig.PutVault(vaultname, vault, "", config)
	if err != nil {
		return "", errors.New(message)
	}
	return
}

func vaultExists(vaultname string, config *configuration.ConfigurationClass) (exists bool, err error) {
	var vaults []serverapi.Vault
	vaults, err = auroraconfig.GetVaultsArray(config)
	if err != nil {
		return false, err
	}

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultname {
			return true, nil
		}
	}

	return false, nil
}

/*
func (this *Vault)getVaults(config *configuration.ConfigurationClass) (err error){
	this.vaults, err = auroraconfig.GetVaultsArray(config)
	if err != nil {
		return err
	}

}

func getVaultIndex(vaultName string, vaults []serverapi.Vault) (vaultIndex int, err error) {
	var found bool
	for i := range vaults {
		if vaults[i].Name == vaultName {
			vaultIndex = i
			found = true
		}
	}
	if found {
		return vaultIndex, nil
	} else {
		err = errors.New("No such vault: " + vaultName)
		return 0, err
	}
}
*/

func appendNoDuplicate(list []string, value string) (newList []string) {
	for i := range list {
		if list[i] == value {
			return list
		}
	}
	newList = append(list, value)
	return newList
}

func getIndex(list []string, value string) (index int, err error) {
	for i := range list {
		if list[i] == value {
			index = i
			return index, nil
		}
	}
	err = errors.New("No such value: " + value)
	return 0, err
}

func remove(list []string, value string) (newList []string, err error) {
	index, err := getIndex(list, value)
	if err != nil {
		return list, err
	}
	newList = append(list[:index], list[index+1:]...)
	return newList, nil
}

func listPermissions(user []string, group []string) (output string, err error) {
	var headers []string = []string{"Users", "Groups"}

	output = printutil.FormatTable(headers, user, group)
	return output, nil
}

func vaults2Exists(vaultName string, vaults []serverapi.Vault) (exists bool) {
	for _, vault := range vaults {
		if vault.Name == vaultName {
			return true
		}
	}
	return false
}

func Rename(vaultName string, newVaultName string, config *configuration.ConfigurationClass) (output string, err error) {

	vaults, err := auroraconfig.GetVaultsArray(config)
	if err != nil {
		return "", err
	}
	if !vaults2Exists(vaultName, vaults) {
		err = errors.New(vaultName + ": No such vault")
		return "", err
	}

	if vaults2Exists(newVaultName, vaults) {
		err = errors.New("Cannot rename to an existing vaultname: " + newVaultName)
		return "", err
	}

	var vault serverapi.Vault
	vault, err = auroraconfig.GetVault(vaultName, config)
	if err != nil {
		return "", err
	}

	vault.Name = newVaultName
	output, err = auroraconfig.PutVault(newVaultName, vault, "", config)
	if err != nil {
		return "", err
	}
	_, err = auroraconfig.DeleteVault(vaultName, config)
	if err != nil {
		return "", err
	}
	return output, nil
}
func Permissions(vaultName string, config *configuration.ConfigurationClass,
	addGroup string, removeGroup string, addUser string, removeUser string) (output string, err error) {

	var vault serverapi.Vault
	vault, err = auroraconfig.GetVault(vaultName, config)
	if err != nil {
		return "", err
	}

	// Do your stuff
	if addGroup != "" {
		vault.Permissions.Groups = appendNoDuplicate(vault.Permissions.Groups, addGroup)
	}

	if addUser != "" {
		vault.Permissions.Users = appendNoDuplicate(vault.Permissions.Users, addUser)
	}

	if removeGroup != "" {
		vault.Permissions.Groups, err = remove(vault.Permissions.Groups, removeGroup)
		if err != nil {
			return "", err
		}
	}

	if removeUser != "" {
		vault.Permissions.Users, err = remove(vault.Permissions.Users, removeUser)
		if err != nil {
			return "", err
		}
	}

	if addGroup == "" && addUser == "" && removeGroup == "" && removeUser == "" {
		// No flags given, list permissions
		output, err = listPermissions(vault.Permissions.Users, vault.Permissions.Groups)
		return output, err
	}

	// Save
	output, err = auroraconfig.PutVault(vaultName, vault, "", config)
	if err != nil {
		return "", err
	}

	return output, nil
}

func ImportVaults(catalogName string, config *configuration.ConfigurationClass) (output string, err error) {
	vaults, err := vaultsFolder2VaultsArray(catalogName)
	if err != nil {
		return "", err
	}

	for _, vault := range vaults {
		output, err = auroraconfig.PutVault(vault.Name, vault, "", config)
		if err != nil {
			return output, err
		}
	}

	return
}

func vaultsFolder2VaultsArray(folderName string) (vaults []serverapi.Vault, err error) {
	folderCount, err := countFolders(folderName)
	if err != nil {
		return nil, err
	}

	vaults = make([]serverapi.Vault, folderCount)

	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		return nil, err
	}
	vaultIndex := 0
	for _, f := range files {
		absolutePath := filepath.Join(folderName, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFolder {
			vaults[vaultIndex], err = secretsFolder2Vault(absolutePath)
			if err != nil {
				return nil, err
			}
			vaultIndex++
		}
	}

	return vaults, nil
}

func secretsFolder2Vault(folderName string) (vault serverapi.Vault, err error) {
	vault.Name = filepath.Base(folderName)
	vault.Secrets = make(map[string]string)
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		return vault, err
	}

	vaultIndex := 0
	for _, f := range files {
		absolutePath := filepath.Join(folderName, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFile {
			// Read file content
			secretContent, err := ioutil.ReadFile(absolutePath)
			if err != nil {
				return vault, err
			}
			// Check for permissions file
			if strings.Contains(filepath.Base(absolutePath), "permission") {
				if !jsonutil.IsLegalJson(string(secretContent)) {
					err = errors.New("Illegal JSON in permissions file " + absolutePath)
					return vault, err
				}
				if len(vault.Permissions.Groups) == 0 && len(vault.Permissions.Users) == 0 {
					vault.Permissions, err = permissionsJson2Permissions(string(secretContent))
					if err != nil {
						return vault, err
					}
				} else {
					err = errors.New("Multiple permission files in a vault is not supported")
					return vault, err
				}
			} else {
				secretContent64 := base64.StdEncoding.EncodeToString(secretContent)
				secretName := filepath.Base(absolutePath)
				vault.Secrets[secretName] = secretContent64
				vaultIndex++
			}
		}
	}

	return
}

func permissionsJson2Permissions(permissionJson string) (permissions serverapi.PermissionsStruct, err error) {
	err = json.Unmarshal([]byte(permissionJson), &permissions)
	if err != nil {
		return permissions, err
	}
	return permissions, err
}

func countFolders(folderName string) (folderCount int, err error) {
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		return 0, err
	}
	folderCount = 0
	for _, f := range files {
		absolutePath := filepath.Join(folderName, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFolder {
			folderCount++
		}
	}
	return folderCount, nil

}
