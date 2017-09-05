package auroraconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

func GetAllVaults(outputFolder string, configuration *configuration.ConfigurationClass) (output string, err error) {
	vaults, err := GetVaultsArray(configuration)
	if err != nil {
		return
	}

	if outputFolder != "" {
		if fileutil.IsLegalFileFolder(outputFolder) == fileutil.SpecIllegal {
			err = errors.New("Illegal file/folder")
			return "", err

		}
	}

	var newline = ""
	for vaultIndex := range vaults {
		content, err := json.Marshal(vaults[vaultIndex])
		contentStr := jsonutil.PrettyPrintJson(string(content))
		output += newline + contentStr
		newline = "\n"
		if outputFolder != "" {
			filename := vaults[vaultIndex].Name + ".json"
			err = fileutil.WriteFile(outputFolder, filename, contentStr)
			if err != nil {
				return "", err
			}
		}
	}

	return output, err
}

func GetVault(vaultname string, configuration *configuration.ConfigurationClass) (vault serverapi.Vault, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname

	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		output, err := serverapi.ResponsItems2MessageString(response)
		if err != nil {
			return vault, err
		}
		err = errors.New(output)
		return vault, err

	}
	vault, err = serverapi.ResponseItems2Vault(response)

	return vault, err
}

func GetVaults(configuration *configuration.ConfigurationClass) (output string, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		if !response.Success {
			output, err := serverapi.ResponsItems2MessageString(response)
			if err != nil {
				return output, err
			}
			err = errors.New(output)
			return output, err

		}

		return output, err

	}

	output, err = serverapi.ResponseItems2Vaults(response)

	return output, err
}

func GetVaultsArray(configuration *configuration.ConfigurationClass) (vaults []serverapi.Vault, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.OpenshiftConfig, configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		if !response.Success {
			output, err := serverapi.ResponsItems2MessageString(response)
			if err != nil {
				return vaults, err
			}
			err = errors.New(output)
			return vaults, err

		}

	}

	vaults, err = serverapi.ResponseItems2VaultsArray(response)

	return vaults, err
}

func GetSecret(vaultName string, secretName string, configuration *configuration.ConfigurationClass) (output string, version string, err error) {
	var vaults []serverapi.Vault
	vaults, err = GetVaultsArray(configuration)

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output = string(decodedSecret)
			version = vaults[vaultindex].Versions[secretName]
		}
	}
	return
}
