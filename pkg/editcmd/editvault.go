package editcmd

import (
	"encoding/json"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

func (editcmd *EditcmdClass) EditVault(vaultname string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	vault, err := auroraconfig.GetVault(vaultname, &editcmd.configuration)
	if err != nil {
		return "", err
	}

	vaultString, err := json.Marshal(vault)
	if err != nil {
		return "", err
	}

	_, output, err = editCycle(string(vaultString), vaultname, "", putVaultString, &editcmd.configuration)

	return output, nil
}

func putVaultString(vaultString string, vaultname string, version string, configuration *configuration.ConfigurationClass) (output string, err error) {
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
