package vault

import (
	"errors"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/printutil"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

/*
type Vault struct {
	vaults []serverapi_v2.Vault
}
*/

func CreateVault(vaultname string, config *configuration.ConfigurationClass) (output string, err error) {
	var vault serverapi_v2.Vault

	exists, err := vaultExists(vaultname, config)
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New("Error: Vault exists")
	}

	vault.Name = vaultname
	vault.Secrets = make(map[string]string)
	//vault.Versions = make(map[string]string)
	//vault.Permissions.Users = make([]string, 0)
	//vault.Permissions.Groups = make([]string, 1)
	//vault.Permissions.Groups[0] = "APP_PaaS_utv"
	message, err := auroraconfig.PutVault(vaultname, vault, "", config)
	if err != nil {
		return "", errors.New(message)
	}
	return
}

func vaultExists(vaultname string, config *configuration.ConfigurationClass) (exists bool, err error) {
	var vaults []serverapi_v2.Vault
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

func getVaultIndex(vaultName string, vaults []serverapi_v2.Vault) (vaultIndex int, err error) {
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

func Permissions(vaultName string, config *configuration.ConfigurationClass,
	addGroup string, removeGroup string, addUser string, removeUser string) (output string, err error) {

	var vault serverapi_v2.Vault
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
