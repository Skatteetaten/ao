package setup

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"github.com/spf13/viper"
	"path/filepath"
)

type SetupClass struct {
	configuration configuration.ConfigurationClass
	initDone      bool
}

func (setupClass *SetupClass) init() (err error) {
	if setupClass.initDone {
		return
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
		if !serverapi.ValidateLogin(setupClass.configuration.GetOpenshiftConfig()) {
			return "", errors.New("Not logged in, please use aoc login")
		}
	}

	error = fileutil.ValidateFileFolderArg(args)
	if error != nil {
		return
	}

	if doSetup {
		error = jsonutil.ValidateOverrides(args, overrideFiles)
	} else {
		error = setupClass.validateImportCommand(args)
	}
	if error != nil {
		return
	}

	var apiEndpoint string
	if doSetup {
		apiEndpoint = "/affiliation/" + setupClass.getAffiliation() + "/setup"
	} else {
		apiEndpoint = "/affiliation/" + setupClass.getAffiliation() + "/auroraconfig"
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
				persistentOptions.Verbose, setupClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
			if err != nil {
				return "", err
			}
		}
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
	if !serverapi.ValidateLogin(setupClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	// Line of code from Mac
	// Line of code from VDI

	return
}

func validateDeploy(args []string) (error error) {
	if len(args) != 0 {
		error = errors.New("Usage: aoc deploy")
	}

	return
}

func (setupClass *SetupClass) getAffiliation() (affiliation string) {
	if setupClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = setupClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}
