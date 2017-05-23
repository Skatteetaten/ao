package auroraconfig

import (
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

func GetContent(filename string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (content string, err error) {
	auroraConfig, err := GetAuroraConfig(persistentOptions, affiliation, openshiftConfig)
	if err != nil {
		return
	}
	var fileFound bool

	for filenameIndex := range auroraConfig.Files {
		if filenameIndex == filename {
			fileFound = true
			content = string(auroraConfig.Files[filenameIndex])
		}
	}
	if !fileFound {
		return "", errors.New("Illegal file/folder")
	}
	return content, nil

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

func GetAuroraConfig(persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (auroraConfig serverapi_v2.AuroraConfig, err error) {
	var apiEndpoint string = "/affiliation/" + affiliation + "/auroraconfig"
	var responses map[string]string

	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug)
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

func PutContent(filename string, content string, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + affiliation + "/auroraconfig/" + filename
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodPut, apiEndpoint, content, persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug)
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
