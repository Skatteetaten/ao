package auroraconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
	"net/http"
	"strconv"
)

const InvalidConfigurationError = "Invalid configuration"

func GetContent(filename string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (content string, version string, err error) {
	auroraConfig, err := GetAuroraConfig(persistentOptions, affiliation, openshiftConfig)
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

func GetAllContent(outputFolder string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (output string, err error) {
	auroraConfig, err := GetAuroraConfig(persistentOptions, affiliation, openshiftConfig)
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

func GetFileList(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (filenames []string, err error) {
	auroraConfig, err := GetAuroraConfig(persistentOptions, affiliation, openshiftConfig)
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

func GetVaults(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (output string, err error) {
	var apiEndpoint string = "/affiliation/" + affiliation + "/vault"
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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
		}

		return output, err

	}

	return
}

func GetVaultsArray(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (vaults []serverapi_v2.Vault, err error) {
	var apiEndpoint string = "/affiliation/" + affiliation + "/vault"
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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
		err = errors.New("Internal error in GetVaults: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return vaults, err
		}
		vaults, err = serverapi_v2.ResponseItems2Vaults(response)
	}

	return vaults, err
}

func GetSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (output string, version string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = GetVaultsArray(persistentOptions, affiliation, openshiftConfig)

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output = string(decodedSecret)
			version = vaults[vaultindex].Versions[secretName]
		}
	}
	return
}

func GetAuroraConfig(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (auroraConfig serverapi_v2.AuroraConfig, err error) {
	var apiEndpoint string = "/affiliation/" + affiliation + "/auroraconfig"
	var responses map[string]string

	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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

func putContent(apiEndpoint string, content string, version string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var responses map[string]string

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	responses, err = serverapi_v2.CallApiWithHeaders(versionHeader, http.MethodPut, apiEndpoint, content, persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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

func PutFile(filename string, content string, version string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + affiliation + "/auroraconfigfile/" + filename

	return putContent(apiEndpoint, content, version, persistentOptions, affiliation, openshiftConfig)
}

func PutSecret(vaultname string, secretname string, secret string, version string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + affiliation + "/vault/" + vaultname + "/secret/" + secretname

	encodedSecret := base64.StdEncoding.EncodeToString([]byte(secret))
	return putContent(apiEndpoint, encodedSecret, version, persistentOptions, affiliation, openshiftConfig)
}

func PutVault(vaultname string, vault serverapi_v2.Vault, version string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + affiliation + "/vault/"

	content, err := json.Marshal(vault)

	return putContent(apiEndpoint, string(content), version, persistentOptions, affiliation, openshiftConfig)

}

func deleteContent(apiEndpoint string, version string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var responses map[string]string

	var versionHeader = make(map[string]string)
	versionHeader["AuroraConfigFileVersion"] = version

	responses, err = serverapi_v2.CallApiWithHeaders(versionHeader, http.MethodDelete, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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

func DeleteVault(vaultname string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + affiliation + "/vault/" + vaultname

	return deleteContent(apiEndpoint, "", persistentOptions, affiliation, openshiftConfig)

}
