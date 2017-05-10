package importcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"io/ioutil"
	"path/filepath"
)

type ImportClass struct {
	configuration configuration.ConfigurationClass
	initDone      bool
}

func validateImportCommand(args []string) (error error) {
	if len(args) != 1 {
		error = errors.New("Usage: aoc import <folder>")
		return
	}
	if fileutil.IsLegalFileFolder(args[0]) != fileutil.SpecIsFolder {
		error = errors.New("Error: " + args[0] + " is not a folder")
	}

	return
}

func (importClass *ImportClass) getAffiliation() (affiliation string) {
	if importClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = importClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (importClass *ImportClass) ExecuteImport(args []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool) (
	output string, error error) {

	//var errorString string

	if !localDryRun {
		if !serverapi.ValidateLogin(importClass.configuration.GetOpenshiftConfig()) {
			return "", errors.New("Not logged in, please use aoc login")
		}
	}

	error = validateImportCommand(args)
	if error != nil {
		return
	}

	var apiEndpoint string
	apiEndpoint = "/affiliation/" + importClass.getAffiliation() + "/auroraconfig"

	var repo = args[0]

	var absolutePath string

	absolutePath, _ = filepath.Abs(repo)

	// Initialize JSON

	jsonStr, err := generateJson(absolutePath,
		importClass.getAffiliation(), persistentOptions.DryRun)
	if err != nil {
		return "", err
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = serverapi.CallApi(apiEndpoint, jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, false, persistentOptions.Localhost,
				persistentOptions.Verbose, importClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
			if err != nil {
				return "", err
			}
		}
	}
	return
}

func generateJson(folder string, affiliation string, dryRun bool) (jsonStr string, error error) {

	var setupCommand jsonutil.SetupCommand
	var auroraConfigPayload jsonutil.AuroraConfigPayload

	var returnMap map[string]json.RawMessage
	var returnMap2 map[string]json.RawMessage
	var secretMap map[string]string = make(map[string]string)

	setupCommand.Affiliation = affiliation

	returnMap, error = jsonutil.JsonFolder2Map(folder, "")
	if error != nil {
		return
	}

	// Loop through all folders and do a JsonFolder2Map on each
	files, _ := ioutil.ReadDir(folder)
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFolder { // Ignore files
			returnMap2, error = jsonutil.JsonFolder2Map(absolutePath, f.Name()+"/")
			returnMap = jsonutil.CombineJsonMaps(returnMap, returnMap2)
		}
	}

	setupCommand.AuroraConfig.Files = returnMap
	setupCommand.AuroraConfig.Secrets = secretMap

	for fileKey := range setupCommand.AuroraConfig.Files {
		secret, err := jsonutil.Json2secretFolder(setupCommand.AuroraConfig.Files[fileKey])
		if err != nil {
			return "", err
		}
		if secret != "" {
			secretMap, err = jsonutil.SecretFolder2Map(secret)
			if err != nil {
				return "", err
			}
			setupCommand.AuroraConfig.Secrets = jsonutil.CombineTextMaps(setupCommand.AuroraConfig.Secrets, secretMap)
		}
	}

	var jsonByte []byte

	auroraConfigPayload.Files = setupCommand.AuroraConfig.Files
	auroraConfigPayload.Secrets = setupCommand.AuroraConfig.Secrets
	jsonByte, error = json.Marshal(auroraConfigPayload)
	if !(error == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling SetupCommand: %v\n", error.Error()))
	}

	jsonStr = string(jsonByte)
	return
}
