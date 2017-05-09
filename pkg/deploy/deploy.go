package deploy

import (
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
)

type DeployClass struct {
	configuration configuration.ConfigurationClass
	initDone      bool
}

func (deployClass *DeployClass) Init() (err error) {
	if deployClass.initDone {
		return
	}
	deployClass.configuration.Init()
	deployClass.initDone = true
	return
}

func (DeployClass *DeployClass) ExecuteDeploy(args []string, overrideFiles []string,
	persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	return
}
