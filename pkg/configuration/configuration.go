package configuration

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/spf13/viper"
)

type ConfigurationClass struct {
	configLocation  string
	openshiftConfig *openshift.OpenshiftConfig
	apiClusterIndex int
	initDone        bool
}

func (configurationClass *ConfigurationClass) init() (err error) {
	if configurationClass.initDone {
		return
	}
	configurationClass.configLocation = viper.GetString("HOME") + "/.aoc.json"
	configurationClass.openshiftConfig, err = openshift.LoadOrInitiateConfigFile(configurationClass.configLocation)
	if err != nil {
		err = errors.New("Error in loading OpenShift configuration")
	}
	// Find index for API cluster,that is the first reachable cluster
	if configurationClass.openshiftConfig != nil {
		for i := range configurationClass.openshiftConfig.Clusters {
			if configurationClass.openshiftConfig.Clusters[i].Reachable {
				configurationClass.apiClusterIndex = i
				break
			}
		}
	}
	configurationClass.initDone = true
	return
}

func (ConfigurationClass *ConfigurationClass) GetOpenshiftConfig() *openshift.OpenshiftConfig {
	ConfigurationClass.init()
	return ConfigurationClass.openshiftConfig
}

func (ConfigurationClass *ConfigurationClass) GetApiClusterIndex() int {
	ConfigurationClass.init()
	return ConfigurationClass.apiClusterIndex
}
