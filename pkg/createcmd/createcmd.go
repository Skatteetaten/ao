package createcmd

import (
	"encoding/base64"
	"errors"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
)

const UsageString = "Usage: aoc create vault <vaultname> | secret <vaultname> <secretname>"

type CreatecmdClass struct {
	configuration configuration.ConfigurationClass
}

func (createcmdClass *CreatecmdClass) getAffiliation() (affiliation string) {
	if createcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = createcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (createcmdClass *CreatecmdClass) createVault(vaultName string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			output = "SECRET"
			for secretindex := range vaults[vaultindex].Secrets {
				output += "\n" + secretindex
			}
		}

	}
	return
}

func (createcmdClass *CreatecmdClass) createSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output += string(decodedSecret)
		}
	}
	return
}

func (createcmdClass *CreatecmdClass) CreateObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, allClusters bool) (output string, err error) {
	err = validateCreatecmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault":
		{

		}
	case "secret":
		{

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
