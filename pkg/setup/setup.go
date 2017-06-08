package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"path/filepath"
	"strings"
)

type SetupClass struct {
	configuration configuration.ConfigurationClass
	initDone      bool
}

/*func (setupClass *SetupClass) init() (err error) {
	if setupClass.initDone {
		return
	}
	setupClass.initDone = true
	return
}*/

func (setupClass *SetupClass) ExecuteSetupImport(args []string, overrideFiles []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool) (
	output string, error error) {

	var errorString string

	//setupClass.init()
	if !localDryRun {
		if !serverapi.ValidateLogin(setupClass.configuration.GetOpenshiftConfig()) {
			return "", errors.New("Not logged in, please use aoc login")
		}
	}

	error = fileutil.ValidateFileFolderArg(args)
	if error != nil {
		return
	}

	error = jsonutil.ValidateOverrides(args, overrideFiles)
	if error != nil {
		return
	}

	var apiEndpoint string
	apiEndpoint = "/affiliation/" + setupClass.getAffiliation() + "/setup"

	var env = args[0]
	var overrideJson []string = args[1:]

	var absolutePath string

	absolutePath, _ = filepath.Abs(env)

	var envFile string      // Filename for app
	var envFolder string    // Short folder name (Env)
	var folder string       // Absolute path of folder
	var parentFolder string // Absolute path of parent

	switch fileutil.IsLegalFileFolder(env) {
	case fileutil.SpecIsFile:
		folder = filepath.Dir(absolutePath)
		envFile = filepath.Base(absolutePath)
	case fileutil.SpecIsFolder:
		folder = absolutePath
		envFile = ""
	}

	parentFolder = filepath.Dir(folder)
	envFolder = filepath.Base(folder)

	if folder == parentFolder {
		errorString += fmt.Sprintf("Application configuration file cannot reside in root directory")
		return "", errors.New(errorString)
	}

	// Initialize JSON

	jsonStr, err := generateJson(envFile, envFolder, folder, parentFolder, overrideJson, overrideFiles,
		setupClass.getAffiliation(), persistentOptions.DryRun)
	if err != nil {
		return "", err
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = serverapi.CallApi(apiEndpoint, jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, false, persistentOptions.Localhost,
				persistentOptions.Verbose, setupClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
			if err != nil {
				return "", err
			}
		}
	}
	return
}

func (setupClass *SetupClass) getAffiliation() (affiliation string) {
	if setupClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = setupClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func generateJson(envFile string, envFolder string, folder string, parentFolder string, overrideJson []string,
	overrideFiles []string, affiliation string, dryRun bool) (jsonStr string, error error) {
	//var apiData ApiInferface
	var setupCommand jsonutil.SetupCommand

	var returnMap map[string]json.RawMessage
	var returnMap2 map[string]json.RawMessage
	var secretMap map[string]string = make(map[string]string)

	setupCommand.SetupParams.Apps = make([]string, 1)
	setupCommand.SetupParams.Envs = make([]string, 1)
	setupCommand.SetupParams.Apps[0] = strings.TrimSuffix(envFile, filepath.Ext(envFile)) //envFile
	setupCommand.SetupParams.Envs[0] = envFolder
	//setupCommand.SetupParams.DryRun = dryRun
	setupCommand.SetupParams.Overrides = jsonutil.Overrides2map(overrideJson, overrideFiles)

	setupCommand.Affiliation = affiliation

	if envFolder != "" {
		returnMap, error = jsonutil.JsonFolder2Map(folder, envFolder+"/")
		if error != nil {
			return
		}
	} else {
		// Import all folders

	}

	returnMap2, error = jsonutil.JsonFolder2Map(parentFolder, "")
	if error != nil {
		return
	}

	setupCommand.AuroraConfig.Files = jsonutil.CombineJsonMaps(returnMap, returnMap2)
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

	for overrideKey := range setupCommand.SetupParams.Overrides {
		secret, err := jsonutil.Json2secretFolder(setupCommand.SetupParams.Overrides[overrideKey])
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

	jsonByte, error = json.Marshal(setupCommand)
	if !(error == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling SetupCommand: %v\n", error.Error()))
	}

	jsonStr = string(jsonByte)
	return
}
