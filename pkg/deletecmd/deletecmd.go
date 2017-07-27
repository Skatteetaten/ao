package deletecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
)

const UsageString = "Usage: aoc delete vault <vaultname> | secret <vaultname> <secretname>"
const vaultDontExistsError = "Error: No such vault"

type DeletecmdClass struct {
	configuration configuration.ConfigurationClass
}

func (deletecmdClass *DeletecmdClass) getAffiliation() (affiliation string) {
	if deletecmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = deletecmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (deletecmdClass *DeletecmdClass) deleteVault(vaultName string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	//var vaults []serverapi_v2.Vault
	//vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
	_, err = auroraconfig.DeleteVault(vaultName, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}
	return
}

func (deletecmdClass *DeletecmdClass) deleteSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	//var vaults []serverapi_v2.Vault
	//vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
	fmt.Println("DEBUG: Delete secret called: " + vaultName + "/" + secretName)
	//vaults, err := auroraconfig.GetVaults(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())

	return
}

func (deletecmdClass *DeletecmdClass) DeleteObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	err = validateDeletecmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault":
		{
			err = deletecmdClass.deleteVault(args[1], persistentOptions)
		}
	case "secret":
		{
			err = deletecmdClass.deleteSecret(args[1], args[2], persistentOptions)
		}
	}
	return

}

func validateDeletecmd(args []string) (err error) {
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
