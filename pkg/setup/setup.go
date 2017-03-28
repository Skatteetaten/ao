package setup

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"github.com/spf13/viper"
	"path/filepath"
)

type SetupClass struct {
	configLocation string
	openshiftConfig *openshift.OpenshiftConfig
	initDone bool
}

func (setupClass *SetupClass) init() (err error) {
	if setupClass.initDone {
		return
	}
	setupClass.configLocation = viper.GetString("HOME") + "/.aoc.json"
	setupClass.openshiftConfig, err = openshift.LoadOrInitiateConfigFile(setupClass.configLocation)
	if err != nil {
		err = errors.New("Error in loading OpenShift configuration")
	}
	return
}

func (setupClass *SetupClass) ExecuteSetup(args []string, overrideFiles []string, persistentOptions *cmdoptions.CommonCommandOptions) (
	output string, error error) {

	var errorString string

	setupClass.init()
	if !persistentOptions.DryRun {
		if !serverapi.ValidateLogin() {
			return "", errors.New("Not logged in, please use aoc login")
		}
	}
	error = validateCommand(args, overrideFiles)
	if error != nil {
		return
	}

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
	jsonStr, err := jsonutil.GenerateJson(envFile, envFolder, folder, parentFolder, overrideJson, overrideFiles, setupClass.openshiftConfig.Affiliation)
	if err != nil {
		return "", err
	} else {
		if persistentOptions.DryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = serverapi.CallApi(jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, false, persistentOptions.Localhost,
				persistentOptions.Verbose)
			if err != nil {
				return "", err
			}
		}
	}
	return
}

func validateCommand(args []string, overrideFiles []string) (error error) {
	var errorString = ""

	if len(args) == 0 {
		errorString += "Missing file/folder "
	} else {
		// Chceck argument 0 for legal file / folder
		validateCode := fileutil.IsLegalFileFolder(args[0])
		if validateCode < 0 {
			errorString += fmt.Sprintf("Illegal file / folder: %v\n", args[0])
		}

		// We have at least one argument, now there should be a correlation between the number of args
		// and the number of override (-f) flags
		if len(overrideFiles) < (len(args) - 1) {
			errorString += fmt.Sprintf("Configuration override specified without file reference flag\n")
		}
		if len(overrideFiles) > (len(args) - 1) {
			errorString += fmt.Sprintf("Configuration overide file reference flag specified without configuration\n")
		}

		// Check for legal JSON argument for each overrideFiles flag
		for i := 1; i < len(args); i++ {
			if !jsonutil.IsLegalJson(args[i]) {
				errorString += fmt.Sprintf("Illegal JSON configuration override: %v\n", args[i])
			}
		}
	}

	if errorString != "" {
		error = errors.New(errorString)
	}
	return
}
