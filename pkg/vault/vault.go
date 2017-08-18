package vault

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

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
