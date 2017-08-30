package importcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type ImportClass struct {
	Configuration *configuration.ConfigurationClass
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

func (importObj *ImportClass) ExecuteImport(args []string, localDryRun bool) (
	output string, error error) {
	//var errorString string

	error = validateImportCommand(args)
	if error != nil {
		return
	}

	auroraConfig, err := auroraconfig.GetAuroraConfig(importObj.Configuration)
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
	apiEndpoint = "/affiliation/" + importObj.Configuration.GetAffiliation() + "/auroraconfig"

	var repo = args[0]

	var absolutePath string
	var responses map[string]string

	absolutePath, _ = filepath.Abs(repo)

	persistentOptions := importObj.Configuration.PersistentOptions
	// Initialize JSON

	jsonStr, err := generateJson(absolutePath,
		importObj.Configuration.GetAffiliation(), persistentOptions.DryRun)
	if err != nil {
		return "", err
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			responses, err = serverapi.CallApi(http.MethodPut, apiEndpoint, jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, true, persistentOptions.Localhost,
				persistentOptions.Verbose, importObj.Configuration.OpenshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
			if err != nil {
				for server := range responses {
					response, err := serverapi.ParseResponse(responses[server])
					if err != nil {
						return "", err
					}
					if !response.Success {
						output, err = serverapi.ResponsItems2MessageString(response)
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
