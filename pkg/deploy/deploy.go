package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

// TODO: Fyll inn envs ved deploy av en app
// TODO: Match apps kun i de envs de finnes.   Innebærer at args må parses to ganger: Først for envs og så for apps

const UsageString = "Usage: deploy <env> <app> <env/app> [--all] [--force] [-e env] [-a app] "

type DeployCommand struct {
	Affiliation string                      `json:"affiliation"`
	SetupParams jsonutil.SetupParamsPayload `json:"setupParams"`
}

type DeployClass struct {
	Configuration *configuration.ConfigurationClass
	setupCommand  DeployCommand
	fuzzyArgs     fuzzyargs.FuzzyArgs
	overrideJsons []string
	auroraConfig  *serverapi.AuroraConfig
}

func (deploy *DeployClass) generateJson(
	affiliation string, dryRun bool) (jsonStr string, err error) {

	applist := deploy.fuzzyArgs.GetApps()

	if len(applist) != 0 {
		deploy.setupCommand.SetupParams.Apps = applist
	} else {
		deploy.setupCommand.SetupParams.Apps = make([]string, 0)
	}
	envlist := deploy.fuzzyArgs.GetEnvs()
	if len(envlist) != 0 {
		deploy.setupCommand.SetupParams.Envs = envlist
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
	persistentOptions *cmdoptions.CommonCommandOptions, localDryRun bool, deployAll bool, force bool, deployVersion string, affiliation string) (output string, err error) {

	if affiliation != "" {
		deploy.Configuration.OpenshiftConfig.Affiliation = affiliation
	}
	ac, err := auroraconfig.GetAuroraConfig(deploy.Configuration)
	if err != nil {
		return "", err
	}
	deploy.auroraConfig = &ac

	err = deploy.validateDeploy(args, applist, envList, deployAll, force)
	if err != nil {
		return "", err
	}

	deploy.overrideJsons = overrideJsons

	if deployVersion != "" {
		err = deploy.updateVersion(deployVersion)
		if err != nil {
			return "", err
		}
	}

	affiliation = deploy.Configuration.GetAffiliation()
	jsonStr, err := deploy.generateJson(affiliation, persistentOptions.DryRun)
	if err != nil {
		return "", err
	}
	var apiEndpoint string = "/affiliation/" + affiliation + "/deploy"
	//var applicationResults []serverapi.ApplicationResult

	if localDryRun {
		return fmt.Sprintf("%v", string(jsonutil.PrettyPrintJson(jsonStr))), nil
	} else {
		var headers map[string]string

		_, err := serverapi.CallDeployWithHeaders(headers, http.MethodPut, apiEndpoint, jsonStr, false, false, persistentOptions.Verbose,
			deploy.Configuration.OpenshiftConfig, persistentOptions.DryRun, persistentOptions.Debug, "", "")

		if err != nil {
			return "", err
		}
		/*for _, response := range responses {



			for applicationResultIndex := range applicationResults {
				out, err := serverapi.ApplicationResult2MessageString(applicationResults[applicationResultIndex])
				if err != nil {
					return out, err
				}
				output += out
			}
		}*/

	}

	return
}

func (deploy *DeployClass) populateFlagsEnvAppList(appList []string, envList []string) (err error) {
	var env string
	var app string

	for i := range appList {
		app, err = deploy.fuzzyArgs.GetFuzzyApp(appList[i])
		if err != nil {
			return err
		}
		if app != "" {
			deploy.fuzzyArgs.AddApp(app)
		} else {
			err = errors.New(appList[i] + ": not found")
			return err
		}
	}

	for i := range envList {
		env, err = deploy.fuzzyArgs.GetFuzzyEnv(envList[i])
		if err != nil {
			return err
		}
		if env != "" {
			deploy.fuzzyArgs.AddEnv(env)
		} else {
			err = errors.New(envList[i] + ": not found")
			return err
		}
	}

	return
}

func (deploy *DeployClass) populateAllAppForEnv(env string) (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(deploy.Configuration)
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
						deploy.fuzzyArgs.AddApp(strings.TrimSuffix(parts[1], ".json"))
					}
				}
			}
		}
	}

	return
}

func (deploy *DeployClass) populateAllEnvForApp(app string) (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(deploy.Configuration)
	if err != nil {
		return err
	}

	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/") {
			// We have a full path name
			parts := strings.Split(filename, "/")
			if strings.Contains(parts[1], app) {
				deploy.fuzzyArgs.AddEnv(parts[0])
			}
		}
	}

	return
}

func (deploy *DeployClass) validateDeploy(args []string, appList []string, envList []string, deployAll bool, force bool) (err error) {
	// We will accept a mixed list of apps, envs and env/app strings and parse them
	// Empty list is illegal

	if len(args) == 0 && len(appList) == 0 && len(envList) == 0 {
		if !deployAll {
			err = errors.New(UsageString)
			return err
		}
	}

	err = deploy.fuzzyArgs.Init(deploy.Configuration)
	if err != nil {
		return err
	}

	if deployAll {
		deploy.fuzzyArgs.DeployAll()
	} else {
		err = deploy.fuzzyArgs.PopulateFuzzyEnvAppList(args, false)
		if err != nil {
			return err
		}

		err = deploy.populateFlagsEnvAppList(appList, envList)
		if err != nil {
			return err
		}
	}

	if len(deploy.fuzzyArgs.GetEnvs()) > 0 && len(deploy.fuzzyArgs.GetApps()) == 0 {
		// User have specified one or more environments, but not an application list, so prefill it
		for i := range deploy.fuzzyArgs.GetEnvs() {
			err := deploy.populateAllAppForEnv(deploy.fuzzyArgs.GetEnvs()[i])
			if err != nil {
				return err
			}
		}
	}

	if len(deploy.fuzzyArgs.GetEnvs()) == 0 && len(deploy.fuzzyArgs.GetApps()) > 0 {
		// User have specified one or more apps, but not an environment list, so prefill it
		for i := range deploy.fuzzyArgs.GetApps() {
			err := deploy.populateAllEnvForApp(deploy.fuzzyArgs.GetApps()[i])
			if err != nil {
				return err
			}
		}
	}

	if len(deploy.fuzzyArgs.GetEnvs()) > 1 || len(deploy.fuzzyArgs.GetApps()) > 1 {
		if !force {
			response, err := executil.PromptYNC(deploy.fuzzyArgs.GetDeploymentSummaryString() + "Are you sure?")
			//			response, err := executil.PromptYNC("This will deploy " + strconv.Itoa(len(deploy.appList)) + " applications in " + strconv.Itoa(len(deploy.envList)) + " environments.  Are you sure?")
			if err != nil {
				return err
			}
			if response != "Y" {
				err = errors.New("Operation cancelled by user")
				return err
			}
		}
	}

	return
}
