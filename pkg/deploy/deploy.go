package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi"
)

type DeployCommand struct {
	Affiliation string                      `json:"affiliation"`
	SetupParams jsonutil.SetupParamsPayload `json:"setupParams"`
}

type DeployClass struct {
	configuration configuration.ConfigurationClass
	initDone      bool
}

func (deployClass *DeployClass) Init() (err error) {
	if deployClass.initDone {
		return
	}
	deployClass.initDone = true
	return
}

func (deployClass *DeployClass) getAffiliation() (affiliation string) {
	if deployClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = deployClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (deployClass *DeployClass) ExecuteDeploy(args []string, overrideFiles []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool) (output string, err error) {

	error := validateDeploy(args)
	if error != nil {
		return
	}
	deployClass.Init()
	if !serverapi.ValidateLogin(deployClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	var env string = args[0]
	var overrideJson []string = args[1:]

	json, error := generateJson(env, "", overrideJson, overrideFiles, "sat", persistentOptions.DryRun)

	var affiliation = deployClass.getAffiliation()

	var apiEndpoint string = "/affiliation/" + affiliation + "/deploy"
	if error != nil {
		return
	} else {
		if localDryRun {
			return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(json))), nil
		} else {
			output, err = serverapi.CallApi(apiEndpoint, json, persistentOptions.ShowConfig,
				persistentOptions.ShowObjects, false, persistentOptions.Localhost,
				persistentOptions.Verbose, deployClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
			if err != nil {
				return "", err
			}
		}
	}

	return
}

func validateDeploy(args []string) (error error) {
	if len(args) != 1 {
		error = errors.New("Usage: aoc deploy <env>")
	}

	return
}

func generateJson(env string, app string, overrideJson []string,
	overrideFiles []string, affiliation string, dryRun bool) (jsonStr string, error error) {
	//var apiData ApiInferface
	var setupCommand DeployCommand

	if app != "" {
		setupCommand.SetupParams.Apps = make([]string, 1)
		setupCommand.SetupParams.Apps[0] = app
	} else {
		setupCommand.SetupParams.Apps = make([]string, 0)
	}
	if env != "" {
		setupCommand.SetupParams.Envs = make([]string, 1)
		setupCommand.SetupParams.Envs[0] = env
	} else {
		setupCommand.SetupParams.Envs = make([]string, 0)
	}

	setupCommand.SetupParams.DryRun = dryRun
	//setupCommand.SetupParams.Overrides = jsonutil.Overrides2map(overrideJson, overrideFiles)
	setupCommand.SetupParams.Overrides = make(map[string]json.RawMessage, 0)
	setupCommand.Affiliation = affiliation

	var jsonByte []byte

	jsonByte, error = json.Marshal(setupCommand)
	if !(error == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling SetupCommand: %v\n", error.Error()))
	}

	jsonStr = string(jsonByte)
	return

}
