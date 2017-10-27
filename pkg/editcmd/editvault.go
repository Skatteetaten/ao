package editcmd

import (
	"encoding/json"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"fmt"
)

func (editcmd *EditcmdClass) EditVault(vaultname string) (string, error) {

	vault, err := auroraconfig.GetVault(vaultname, editcmd.Configuration)
	if err != nil {
		return "", err
	}

	vaultString, err := json.Marshal(vault)
	if err != nil {
		return "", err
	}

	debug := editcmd.Configuration.PersistentOptions.Debug

	onSave := func(modified string) ([]string, error) {
		return putVaultString(modified, "", editcmd.Configuration)
	}

	_, output, err := editCycle(string(vaultString), vaultname, debug, onSave)
	fmt.Println(err)

	return output, nil
}

func putVaultString(vaultString string, version string, configuration *configuration.ConfigurationClass) ([]string, error) {
	var vault auroraconfig.Vault

	err := json.Unmarshal([]byte(vaultString), &vault)
	if err != nil {
		return []string{}, err
	}
	output, _, err := auroraconfig.PutVault(vault, version, configuration)
	if err != nil {
		return []string{}, err
	}
	return []string{output}, err
}
