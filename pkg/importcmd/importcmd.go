package importcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type ImportClass struct {
	configuration configuration.ConfigurationClass
}

func (importObj *ImportClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	importObj.configuration.Init(persistentOptions)
	return
}

func validateImportCommand(args []string) (error error) {
	if len(args) != 1 {
		error = errors.New("Usage: import <folder>")
		return
	}
	if fileutil.IsLegalFileFolder(args[0]) != fileutil.SpecIsFolder {
		error = errors.New("Error: " + args[0] + " is not a folder")
	}

	return
}

func (importObj *ImportClass) ExecuteImport(args []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool) (
	output string, error error) {

	importObj.init(persistentOptions)
	//var errorString string

	if !localDryRun {
		if !serverapi_v2.ValidateLogin(importObj.configuration.GetOpenshiftConfig()) {
			return "", errors.New("Not logged in, please use ao login")
		}
	}

	error = validateImportCommand(args)
	if error != nil {
		return
	}

	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, importObj.configuration.GetAffiliation(), importObj.configuration.GetOpenshiftConfig())
	if err != nil {
		return "", err
	}

	// We allow just one file to be in the config, that will be the global about.json that is illegal to delete.
	// Will probably be safe to overwrite
	if len(auroraConfig.Files) > 1 {
		err = errors.New("Import not allowed into a non-empty AuroraConfig as it will overwrite the config.")
		return "", err
	}
	var apiEndpoint string
	apiEndpoint = "/affiliation/" + importObj.configuration.GetAffiliation() + "/auroraconfig"

	var repo = args[0]

	var absolutePath string
	var responses map[string]string

	absolutePath, _ = filepath.Abs(repo)

	// Initialize JSON

	jsonStr, err := generateJson(absolutePath,
		importObj.configuration.GetAffiliation(), persistentOptions.DryRun)
	if err != nil {
		return "", err
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			responses, err = serverapi_v2.CallApi(http.MethodPut, apiEndpoint, jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, true, persistentOptions.Localhost,
				persistentOptions.Verbose, importObj.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
			if err != nil {
				for server := range responses {
					response, err := serverapi_v2.ParseResponse(responses[server])
					if err != nil {
						return "", err
					}
					if !response.Success {
						output, err = serverapi_v2.ResponsItems2MessageString(response)
					}
				}
				return output, nil
			}
		}
	}
	return
}

func generateJson(folder string, affiliation string, dryRun bool) (jsonStr string, err error) {

	var setupCommand jsonutil.SetupCommand
	var auroraConfigPayload jsonutil.AuroraConfigPayload

	var returnMap map[string]json.RawMessage
	var returnMap2 map[string]json.RawMessage

	setupCommand.Affiliation = affiliation

	returnMap, err = jsonutil.JsonFolder2Map(folder, "")
	if err != nil {
		return
	}

	// Loop through all folders and do a JsonFolder2Map on each
	files, _ := ioutil.ReadDir(folder)
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFolder { // Ignore files
			returnMap2, err = jsonutil.JsonFolder2Map(absolutePath, f.Name()+"/")
			if err != nil {
				return "", err
			}
			returnMap = jsonutil.CombineJsonMaps(returnMap, returnMap2)
		}
	}

	setupCommand.AuroraConfig.Files = returnMap

	var jsonByte []byte

	auroraConfigPayload.Files = setupCommand.AuroraConfig.Files
	jsonByte, err = json.Marshal(auroraConfigPayload)
	if !(err == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling SetupCommand: %v\n", err.Error()))
	}

	jsonStr = string(jsonByte)
	return
}
