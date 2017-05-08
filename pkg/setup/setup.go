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
	configLocation  string
	openshiftConfig *openshift.OpenshiftConfig
	apiClusterIndex int
	initDone        bool
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
	// Find index for API cluster,that is the first reachable cluster
	if setupClass.openshiftConfig != nil {
		for i := range setupClass.openshiftConfig.Clusters {
			if setupClass.openshiftConfig.Clusters[i].Reachable {
				setupClass.apiClusterIndex = i
				break
			}
		}
	}
	setupClass.initDone = true
	return
}

func (setupClass *SetupClass) getApiCluster() *openshift.OpenshiftCluster {
	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	openshiftConfig, err := openshift.LoadOrInitiateConfigFile(configLocation)
	if err != nil {
		fmt.Println("Error in loading OpenShift configuration")
		return nil
	}
	for i := range openshiftConfig.Clusters {
		if openshiftConfig.Clusters[i].Reachable {
			return openshiftConfig.Clusters[i]
		}
	}
	return nil
}

func (setupClass *SetupClass) validateImportCommand(args []string) (error error) {
	error = setupClass.validateFileFolderArg(args)
	if error != nil {
		return
	}

	if len(args) > 1 {
		error = errors.New("Usage: aoc import file | folder")
		return
	}
	return
}

func (setupClass *SetupClass) ExecuteSetupImport(args []string, overrideFiles []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool, doSetup bool) (
	output string, error error) {

	var errorString string

	setupClass.init()
	if !localDryRun {
		if !serverapi.ValidateLogin(setupClass.openshiftConfig) {
			return "", errors.New("Not logged in, please use aoc login")
		}
	}

	if doSetup {
		error = setupClass.validateSetupCommand(args, overrideFiles)
	} else {
		error = setupClass.validateImportCommand(args)
	}
	if error != nil {
		return
	}

	var apiEndpoint string
	if doSetup {
		apiEndpoint = "/setup"
	} else {
		apiEndpoint = "/auroraconfig/" + setupClass.getAffiliation()
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

	jsonStr, err := jsonutil.GenerateJson(envFile, envFolder, folder, parentFolder, overrideJson, overrideFiles,
		setupClass.getAffiliation(), persistentOptions.DryRun, doSetup)
	if err != nil {
		return "", err
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = serverapi.CallApi(apiEndpoint, jsonStr, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, false, persistentOptions.Localhost,
				persistentOptions.Verbose, setupClass.openshiftConfig, persistentOptions.DryRun, persistentOptions.Debug)
			if err != nil {
				return "", err
			}
		}
	}
	return
}

func (setupClass *SetupClass) validateSetupCommand(args []string, overrideFiles []string) (error error) {
	var errorString = ""

	error = setupClass.validateFileFolderArg(args)
	if error != nil {
		return
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

	if errorString != "" {
		error = errors.New(errorString)
	}
	return
}

func (setupClass *SetupClass) ExecuteDeploy(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (
	output string, error error) {

	error = validateDeploy(args)
	if error != nil {
		return
	}

	setupClass.init()
	if !serverapi.ValidateLogin(setupClass.openshiftConfig) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	// Line of code from Mac
	// Line of code from VDI

	l
	return
}

func validateDeploy(args []string) (error error) {
	if len(args) != 0 {
		error = errors.New("Usage: aoc deploy")
	}

	return
}

func (SetupClass *SetupClass) validateFileFolderArg(args []string) (error error) {
	var errorString string

	if len(args) == 0 {
		errorString += "Missing file/folder "
	} else {
		// Chceck argument 0 for legal file / folder
		validateCode := fileutil.IsLegalFileFolder(args[0])
		if validateCode < 0 {
			errorString += fmt.Sprintf("Illegal file / folder: %v\n", args[0])
		}

	}

	if errorString != "" {
		return errors.New(errorString)
	}
	return

}

func (setupClass *SetupClass) getAffiliation() (affiliation string) {
	if setupClass.openshiftConfig != nil {
		affiliation = setupClass.openshiftConfig.Affiliation
	}
	return
}
