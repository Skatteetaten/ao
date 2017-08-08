package createcmd

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

const UsageString = "Usage: create vault <vaultname> | secret <vaultname> <secretname>"
const vaultExistsError = "Error: Vault exists"

type CreatecmdClass struct {
	configuration configuration.ConfigurationClass
}

func (createcmd *CreatecmdClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	createcmd.configuration.Init(persistentOptions)
	return
}

func (createcmdClass *CreatecmdClass) vaultExists(vaultname string, persistentOptions *cmdoptions.CommonCommandOptions) (exists bool, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaultsArray(persistentOptions, createcmdClass.configuration.GetAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
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

func (createcmdClass *CreatecmdClass) createVault(vaultname string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var vault serverapi_v2.Vault

	exists, err := createcmdClass.vaultExists(vaultname, persistentOptions)
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New(vaultExistsError)
	}

	vault.Name = vaultname
	vault.Secrets = make(map[string]string)
	//vault.Versions = make(map[string]string)
	//vault.Permissions.Users = make([]string, 0)
	//vault.Permissions.Groups = make([]string, 1)
	//vault.Permissions.Groups[0] = "APP_PaaS_utv"
	message, err := auroraconfig.PutVault(vaultname, vault, "", persistentOptions, createcmdClass.configuration.GetAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return "", errors.New(message)
	}
	return
}

func (createcmdClass *CreatecmdClass) createSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	//var vaults []serverapi_v2.Vault
	//vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())

	return "Not implemented yet, use edit secret to create a new secret", nil
}

func (createcmd *CreatecmdClass) CreateObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, allClusters bool) (output string, err error) {
	createcmd.init(persistentOptions)
	if !serverapi_v2.ValidateLogin(createcmd.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use ao login")
	}

	err = validateCreatecmd(args)

	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault":
		{
			output, err = createcmd.createVault(args[1], persistentOptions)
		}
	case "secret":
		{
			output, err = createcmd.createSecret(args[1], args[2], persistentOptions)
		}
	}
	return

}

func validateCreatecmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(UsageString)
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault":
		{
			if len(args) != 2 {
				err = errors.New(UsageString)
				return
			}
		}
	case "secret":
		{
			if len(args) != 3 {
				err = errors.New(UsageString)
				return
			}
		}
	}
	return

}
