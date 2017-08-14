package deploy

import (
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

// Function to update the version before deploy
//

func (deploy *DeployClass) updateVersion(deployVersion string) (err error) {
	// Check for a single app deploy
	if len(deploy.appList) != 1 {
		err := errors.New("Setting version is only allowed in single-app deploys")
		return err
	}
	if len(deploy.envList) != 1 {
		err := errors.New("Setting version is only allowed in single-env deploys")
		return err
	}

	if deploy.auroraConfig == nil {
		auroraConfig, err := auroraconfig.GetAuroraConfig(&deploy.configuration)
		if err != nil {
			return err
		}
		deploy.auroraConfig = &auroraConfig
	}

	configfilename := deploy.envList[0] + "/" + deploy.appList[0] + ".json"
	for filename := range deploy.auroraConfig.Files {

	}

}
