package auroraconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"net/http"
	"strconv"
)

const InvalidConfigurationError = "Invalid configuration"

func GetContent(filename string, configuration *configuration.ConfigurationClass) (content string, version string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}
	var fileFound bool = false

	_, fileFound = auroraConfig.Files[filename]
	if fileFound {
		content = string(auroraConfig.Files[filename])
	}

	version = auroraConfig.Versions[filename]

	if !fileFound {
		return "", "", errors.New("Illegal file/folder")
	}
	return content, version, nil

}

func GetAllContent(outputFolder string, configuration *configuration.ConfigurationClass) (output string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}

	if outputFolder != "" {
		if fileutil.IsLegalFileFolder(outputFolder) == fileutil.SpecIllegal {
			err = errors.New("Illegal file/folder")
			return "", err

		}
		var content string
		for filename := range auroraConfig.Files {
			content = jsonutil.PrettyPrintJson(string(auroraConfig.Files[filename]))
			err = fileutil.WriteFile(outputFolder, filename, content)
			if err != nil {
				return "", err
			}
		}
	}

	outputBytes, err := json.Marshal(auroraConfig)
	output = jsonutil.PrettyPrintJson(string(outputBytes))
	return output, err

}

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

func GetFileList(configuration *configuration.ConfigurationClass) (filenames []string, err error) {
	auroraConfig, err := GetAuroraConfig(configuration)
	if err != nil {
		return
	}
	filenames = make([]string, len(auroraConfig.Files))

	var filenameIndex = 0
	for filename := range auroraConfig.Files {
		filenames[filenameIndex] = filename
		filenameIndex++
	}
	return filenames, nil
}

// Deprecated when Secrets are removed from AuroraConfig
/*func GetSecretList(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (secretnames []string, err error) {
	auroraConfig, err := GetAuroraConfig(persistentOptions, affiliation, openshiftConfig)
	if err != nil {
		return
	}
	secretnames = make([]string, len(auroraConfig.Secrets))

	var secretnameIndex = 0
	for secretname := range auroraConfig.Secrets {
		secretnames[secretnameIndex] = secretname
		secretnameIndex++
	}
	return secretnames, nil
}*/

func GetVault(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (vault serverapi_v2.Vault, err error) {

	return
}

func GetVaults(configuration *configuration.ConfigurationClass) (output string, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return output, err
			}
			if !response.Success {
				output, err := serverapi_v2.ResponsItems2MessageString(response)
				if err != nil {
					return output, err
				}
				err = errors.New(output)
				return output, err

			}
			output = responses[server]
			return output, err
		}

		return output, err

	}

	if len(responses) != 1 {
		err = errors.New("Internal error in GetVaults: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return output, err
		}
		output, err = serverapi_v2.ResponseItems2Vaults(response)
	}

	return output, err

	return
}

func GetVaultsArray(configuration *configuration.ConfigurationClass) (vaults []serverapi_v2.Vault, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return vaults, err
			}
			if !response.Success {
				output, err := serverapi_v2.ResponsItems2MessageString(response)
				if err != nil {
					return vaults, err
				}
				err = errors.New(output)
				return vaults, err

			}
		}

		return vaults, err

	}

	if len(responses) != 1 {
		err = errors.New("Internal error in GetVaultsArray: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return vaults, err
		}
		vaults, err = serverapi_v2.ResponseItems2VaultsArray(response)
	}

	return vaults, err
}

func GetSecret(vaultName string, secretName string, configuration *configuration.ConfigurationClass) (output string, version string, err error) {
	var vaults []serverapi_v2.Vault
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

func GetAuroraConfig(configuration *configuration.ConfigurationClass) (auroraConfig serverapi_v2.AuroraConfig, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfig"
	var responses map[string]string

	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects, true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return auroraConfig, err
			}
			if !response.Success {
				output, err := serverapi_v2.ResponsItems2MessageString(response)
				if err != nil {
					return auroraConfig, err
				}
				err = errors.New(output)
				return auroraConfig, err

			}
		}

		return auroraConfig, err
	}

	if len(responses) != 1 {
		err = errors.New("Internal error in GetContent: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return auroraConfig, err
		}
		auroraConfig, err = serverapi_v2.ResponseItems2AuroraConfig(response)

	}

	return auroraConfig, nil
}

func PutAuroraConfig(auroraConfig serverapi_v2.AuroraConfig, configuration *configuration.ConfigurationClass) (err error) {
	content, err := json.Marshal(auroraConfig)
	if err != nil {
		return err
	}

	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfig"

	_, err = putContent(apiEndpoint, string(content), "", configuration)
	if err != nil {
		return err
	}
	return
}

func putContent(apiEndpoint string, content string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var responses map[string]string

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	/* headers map[string]string, httpMethod string, apiEndpoint string, combindedJson string, api bool, localhost bool, verbose bool,
	openshiftConfig *openshift.OpenshiftConfig, dryRun bool, debug bool, apiAddress string, token string*/

	responses, err = serverapi_v2.CallApiWithHeaders(versionHeader, http.MethodPut, apiEndpoint, content, true,
		configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose,
		configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun, configuration.GetPersistentOptions().Debug,
		configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)

	/*
		configuration.GetPersistentOptions().ShowConfig,
		configuration.GetPersistentOptions().ShowObjects,


		true, configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose, configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun,
		configuration.GetPersistentOptions().Debug, configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)*/
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return "", err
			}
			if !response.Success {
				validationMessages, _ := serverapi_v2.ResponsItems2MessageString(response)
				return validationMessages, errors.New(InvalidConfigurationError)
			}
		}

	}
	return
}

func PutFile(filename string, content string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/auroraconfigfile/" + filename

	return putContent(apiEndpoint, content, version, configuration)
}

func PutSecret(vaultname string, secretname string, secret string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname + "/secret/" + secretname

	encodedSecret := base64.StdEncoding.EncodeToString([]byte(secret))
	return putContent(apiEndpoint, encodedSecret, version, configuration)
}

func PutVault(vaultname string, vault serverapi_v2.Vault, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/"

	content, err := json.Marshal(vault)

	return putContent(apiEndpoint, string(content), version, configuration)

}

func deleteContent(apiEndpoint string, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var responses map[string]string

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	responses, err = serverapi_v2.CallApiWithHeaders(versionHeader, http.MethodPut, apiEndpoint, "", true,
		configuration.GetPersistentOptions().Localhost,
		configuration.GetPersistentOptions().Verbose,
		configuration.GetOpenshiftConfig(), configuration.GetPersistentOptions().DryRun, configuration.GetPersistentOptions().Debug,
		configuration.GetPersistentOptions().ServerApi, configuration.GetPersistentOptions().Token)

	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return "", err
			}
			if !response.Success {
				validationMessages, _ := serverapi_v2.ResponsItems2MessageString(response)
				return validationMessages, errors.New(validationMessages)
			}
		}

	}
	return
}

func DeleteVault(vaultname string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname

	return deleteContent(apiEndpoint, "", configuration)

}
