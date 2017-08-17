package deploy

import (
	"encoding/json"

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
		auroraConfig, err := auroraconfig.GetAuroraConfig(deploy.Configuration)
		if err != nil {
			return err
		}
		deploy.auroraConfig = &auroraConfig
	}

	configfilename := deploy.envList[0] + "/" + deploy.appList[0] + ".json"
	for filename := range deploy.auroraConfig.Files {
		if filename == configfilename {
			deploy.auroraConfig.Files[filename], err = setVersion(deploy.auroraConfig.Files[filename], deployVersion)
			if err != nil {
				return err
			}
		}
	}

	err = auroraconfig.PutAuroraConfig(*deploy.auroraConfig, deploy.Configuration)
	if err != nil {
		return err
	}
	return

}

func getVersion(configFile json.RawMessage) (version string, err error) {
	var configFileInterface interface{}

	err = json.Unmarshal(configFile, &configFileInterface)
	if err != nil {
		return "", err
	}

	configFileMap := configFileInterface.(map[string]interface{})

	version = string(configFileMap["version"].(string))

	return
}

func setVersion(configFile json.RawMessage, version string) (updatedConfigFile json.RawMessage, err error) {
	var configFileInterface interface{}

	err = json.Unmarshal(configFile, &configFileInterface)
	if err != nil {
		return nil, err
	}

	configFileMap := configFileInterface.(map[string]interface{})

	configFileMap["version"] = version

	updatedConfigFile, err = json.Marshal(configFileMap)
	if err != nil {
		return nil, err
	}

	return

}
