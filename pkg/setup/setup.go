package setup

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/boober"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/spf13/viper"
	"path/filepath"
)

func ExecuteSetup(args []string, overrideFiles []string, persistentOptions *cmdoptions.CommonCommandOptions) (
	output string, error error) {

	var errorString string
	var affiliation string

	if !persistentOptions.DryRun {
		if !boober.ValidateLogin() {
			return "", errors.New("Not logged in, please use aoc login")
		}
		affiliation, error = GetAffiliation()
		if error != nil {
			return
		}
	}
	error = validateCommand(args, overrideFiles)
	if error != nil {
		return
	}

	var absolutePath string

	absolutePath, _ = filepath.Abs(args[0])

	var envFile string      // Filename for app
	var envFolder string    // Short folder name (Env)
	var folder string       // Absolute path of folder
	var parentFolder string // Absolute path of parent

	switch fileutil.IsLegalFileFolder(args[0]) {
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
	jsonStr, err := jsonutil.GenerateJson(envFile, envFolder, folder, parentFolder, args, overrideFiles, affiliation)
	if err != nil {
		return "", err
	} else {
		if persistentOptions.DryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = boober.CallBoober(jsonStr, persistentOptions.ShowConfig,
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

func GetAffiliation() (string, error) {
	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	openshiftConfig, err := openshift.LoadOrInitiateConfigFile(configLocation)
	if err != nil {
		return "", errors.New("Error in loading OpenShift configuration")
	}
	return openshiftConfig.Affiliation, nil
}
