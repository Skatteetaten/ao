package deploy

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/serverapi"
)

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

func (deployClass *DeployClass) ExecuteDeploy(args []string, overrideFiles []string,
	persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {

	error := validateDeploy(args)
	if error != nil {
		return
	}
	deployClass.Init()
	if !serverapi.ValidateLogin(deployClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}
	return
}

func validateDeploy(args []string) (error error) {
	if len(args) != 0 {
		error = errors.New("Usage: aoc deploy")
	}

	return
}
