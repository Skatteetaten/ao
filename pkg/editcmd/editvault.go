package editcmd

import (
	"encoding/json"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

func (editcmd *EditcmdClass) EditVault(vaultname string) (output string, err error) {
	vault, err := auroraconfig.GetVault(vaultname, editcmd.Configuration)
	if err != nil {
		return "", err
	}

	vaultString, err := json.Marshal(vault)
	if err != nil {
		return "", err
	}

	_, output, err = editCycle(string(vaultString), vaultname, "", putVaultString, editcmd.Configuration)

	return output, nil
}

func putVaultString(vaultname string, vaultString string, version string, configuration *configuration.ConfigurationClass) (output string, err error) {
	var vault serverapi_v2.Vault

	err = json.Unmarshal([]byte(vaultString), &vault)
	if err != nil {
		return "", err
	}
	output, err = auroraconfig.PutVault(vaultname, vault, version, configuration)
	if err != nil {
		return "", err
	}
	return output, err
}
