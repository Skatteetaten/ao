package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/executil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
	"net/http"
	"strconv"
	"strings"
)

const UsageString = "Usage: deploy <env> <app> <env/app> [--all] [--force] [-e env] [-a app] "

type DeployCommand struct {
	Affiliation string                      `json:"affiliation"`
	SetupParams jsonutil.SetupParamsPayload `json:"setupParams"`
}

type DeployClass struct {
	configuration configuration.ConfigurationClass
	setupCommand  DeployCommand
	appList       []string
	envList       []string
	legalAppList  []string
	legalEnvList  []string
	overrideJsons []string
}

func (deploy *DeployClass) addLegalApp(app string) {
	for i := range deploy.legalAppList {
		if deploy.legalAppList[i] == app {
			return
		}
	}
	deploy.legalAppList = append(deploy.legalAppList, app)
	return
}

func (deploy *DeployClass) addLegalEnv(env string) {
	for i := range deploy.legalEnvList {
		if deploy.legalEnvList[i] == env {
			return
		}
	}
	deploy.legalEnvList = append(deploy.legalEnvList, env)
	return
}

func (deploy *DeployClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	deploy.configuration.Init(persistentOptions)
	return
}

func (deploy *DeployClass) generateJson(
	affiliation string, dryRun bool) (jsonStr string, err error) {

	if len(deploy.appList) != 0 {
		deploy.setupCommand.SetupParams.Apps = deploy.appList
	} else {
		deploy.setupCommand.SetupParams.Apps = make([]string, 0)
	}
	if len(deploy.envList) != 0 {
		deploy.setupCommand.SetupParams.Envs = deploy.envList
	} else {
		deploy.setupCommand.SetupParams.Envs = make([]string, 0)
	}

	//setupCommand.SetupParams.DryRun = dryRun
	deploy.setupCommand.SetupParams.Overrides, err = jsonutil.OverrideJsons2map(deploy.overrideJsons)
	if err != nil {
		return "", err
	}
	deploy.setupCommand.Affiliation = affiliation

	var jsonByte []byte

	jsonByte, err = json.Marshal(deploy.setupCommand)
	if !(err == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling SetupCommand: %v\n", err.Error()))
	}

	jsonStr = string(jsonByte)
	return

}

func (deploy *DeployClass) ExecuteDeploy(args []string, overrideJsons []string, applist []string, envList []string,
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool, deployAll bool, force bool) (output string, err error) {

	deploy.init(persistentOptions)
	if !serverapi_v2.ValidateLogin(deploy.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	err = deploy.validateDeploy(args, applist, envList, deployAll, force)
	if err != nil {
		return "", err
	}

	var affiliation = deploy.configuration.GetAffiliation()
	json, err := deploy.generateJson(affiliation, persistentOptions.DryRun)
	if err != nil {
		return "", err
	}
	var apiEndpoint string = "/affiliation/" + affiliation + "/deploy"
	var responses map[string]string
	var applicationResults []serverapi_v2.ApplicationResult

	if localDryRun {
		return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(json))), nil
	} else {
		responses, err = serverapi_v2.CallApi(http.MethodPut, apiEndpoint, json, persistentOptions.ShowConfig,
			persistentOptions.ShowObjects, false, persistentOptions.Localhost,
			persistentOptions.Verbose, deploy.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug, persistentOptions.ServerApi, persistentOptions.Token)
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
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return "", err
			}
			if response.Success {
				applicationResults, err = serverapi_v2.ResponseItems2ApplicationResults(response)
			}
			for applicationResultIndex := range applicationResults {
				out, err := serverapi_v2.ApplicationResult2MessageString(applicationResults[applicationResultIndex])
				if err != nil {
					return out, err
				}
				output += out
			}
		}

	}

	return
}

func (deploy *DeployClass) getLegalEnvAppList() (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(deploy.configuration.GetPersistentOptions(), deploy.configuration.GetAffiliation(), deploy.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/") {
			// We have a full path name
			parts := strings.Split(filename, "/")
			deploy.addLegalEnv(parts[0])
			if !strings.Contains(parts[1], "about.json") {
				if strings.HasSuffix(parts[1], ".json") {
					deploy.addLegalApp(strings.TrimSuffix(parts[1], ".json"))
				}

			}
		}
	}

	return
}

// Try to match an argument with an app, returns "" if none found
func (deploy *DeployClass) getFuzzyApp(arg string) (app string, err error) {
	// First check for exact match
	for i := range deploy.legalAppList {
		if deploy.legalAppList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an app name that contains the string
	for i := range deploy.legalAppList {
		if strings.Contains(deploy.legalAppList[i], arg) {
			if app != "" {
				err = errors.New(arg + ": Not a unique application identifier, matching " + app + " and " + deploy.legalAppList[i])
				return "", err
			}
			app = deploy.legalAppList[i]
		}
	}
	return app, nil
}

// Try to match an argument with an env, returns "" if none found
func (deploy *DeployClass) getFuzzyEnv(arg string) (env string, err error) {
	// First check for exact match
	for i := range deploy.legalEnvList {
		if deploy.legalEnvList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an env name that contains the string
	for i := range deploy.legalEnvList {
		if strings.Contains(deploy.legalEnvList[i], arg) {
			if env != "" {
				err = errors.New(arg + ": Not a unique environment identifier, matching both " + env + " and " + deploy.legalEnvList[i])
				return "", err
			}
			env = deploy.legalEnvList[i]
		}
	}
	return env, nil
}

func (deploy *DeployClass) populateFuzzyEnvAppList(args []string) (err error) {

	for i := range args {
		var env string
		var app string

		if strings.Contains(args[i], "/") {
			parts := strings.Split(args[i], "/")
			env, err = deploy.getFuzzyEnv(parts[0])
			if err != nil {
				return err
			}
			app, err = deploy.getFuzzyApp(parts[1])
			if err != nil {
				return err
			}
		} else {
			env, err = deploy.getFuzzyEnv(args[i])
			if err != nil {
				return err
			}
			app, err = deploy.getFuzzyApp(args[i])
			if err != nil {
				return err
			}
			if env != "" && app != "" {
				err = errors.New(args[i] + ": Not a unique identifier, matching both environment " + env + " and application " + app)
				return err
			}
		}
		if env == "" && app == "" {
			// None found, return error
			err = errors.New(args[i] + ": not found")
			return err
		}
		if env != "" {
			deploy.envList = append(deploy.envList, env)
		}
		if app != "" {
			deploy.appList = append(deploy.appList, app)
		}

	}
	return
}

func (deploy *DeployClass) populateFlagsEnvAppList(appList []string, envList []string) (err error) {
	var env string
	var app string

	for i := range appList {
		app, err = deploy.getFuzzyApp(appList[i])
		if err != nil {
			return err
		}
		if app != "" {
			deploy.appList = append(deploy.appList, app)
		} else {
			err = errors.New(appList[i] + ": not found")
			return err
		}
	}

	for i := range envList {
		env, err = deploy.getFuzzyEnv(envList[i])
		if err != nil {
			return err
		}
		if env != "" {
			deploy.envList = append(deploy.envList, env)
		} else {
			err = errors.New(envList[i] + ": not found")
			return err
		}
	}

	return
}

func (deploy *DeployClass) populateAllAppForEnv(env string) (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(deploy.configuration.GetPersistentOptions(), deploy.configuration.GetAffiliation(), deploy.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/") {
			// We have a full path name
			parts := strings.Split(filename, "/")
			if parts[0] == env {
				if !strings.Contains(parts[1], "about.json") {
					if strings.HasSuffix(parts[1], ".json") {
						deploy.appList = append(deploy.appList, strings.TrimSuffix(parts[1], ".json"))
					}
				}
			}
		}
	}

	return
}

func (deploy *DeployClass) populateAll() {
	deploy.envList = deploy.legalEnvList
	deploy.appList = deploy.legalAppList
	return
}

func (deploy *DeployClass) validateDeploy(args []string, appList []string, envList []string, deployAll bool, force bool) (err error) {
	// We will accept a mixed list of apps, envs and env/app strings and parse them
	// Empty list is illegal

	if len(args) == 0 {
		if !deployAll {
			err = errors.New(UsageString)
			return err
		}
	}

	err = deploy.getLegalEnvAppList()
	if err != nil {
		return err
	}

	if deployAll {
		deploy.populateAll()
		if !force {

			response, err := executil.PromptYNC("This will deploy " + strconv.Itoa(len(deploy.appList)) + " applications in " + strconv.Itoa(len(deploy.envList)) + " environments.  Are you sure?")
			if err != nil {
				return err
			}
			if response != "Y" {
				err = errors.New("Operation cancelled by user")
				return err
			}

		}

	} else {
		err = deploy.populateFuzzyEnvAppList(args)
		if err != nil {
			return err
		}

		err = deploy.populateFlagsEnvAppList(appList, envList)
		if err != nil {
			return err
		}
	}

	if len(deploy.envList) > 0 && len(deploy.appList) == 0 {
		// User have specified one or more environments, but not an application list, so prefill it
		for i := range deploy.envList {
			err := deploy.populateAllAppForEnv(deploy.envList[i])
			if err != nil {
				return err
			}
		}
		response, err := executil.PromptYNC("This will deploy " + strconv.Itoa(len(deploy.appList)) + " applications in " + strconv.Itoa(len(deploy.envList)) + " environments.  Are you sure?")
		if err != nil {
			return err
		}
		if response != "Y" {
			err = errors.New("Operation cancelled by user")
			return err
		}
	}

	return
}
