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
	apiClusterName  string
	initDone        bool
}

func (configurationClass *ConfigurationClass) init() (err error) {
	if configurationClass.initDone {
		return
	}
	configurationClass.configLocation = viper.GetString("HOME") + "/.aoc.json"
	configurationClass.openshiftConfig, err = openshift.LoadOrInitiateConfigFile(configurationClass.configLocation, false)
	if err != nil {
		err = errors.New("Error in loading OpenShift configuration")
	}
	// Find index for API cluster,that is the first reachable cluster
	if configurationClass.openshiftConfig != nil {
		for i := range configurationClass.openshiftConfig.Clusters {
			if configurationClass.openshiftConfig.Clusters[i].Name == configurationClass.openshiftConfig.APICluster {
				configurationClass.apiClusterIndex = i
				break
			}
		}
	}
	configurationClass.initDone = true
	return
}

func (configurationClass *ConfigurationClass) GetOpenshiftConfig() *openshift.OpenshiftConfig {
	configurationClass.init()
	return configurationClass.openshiftConfig
}

func (configurationClass *ConfigurationClass) GetApiClusterIndex() int {
	configurationClass.init()
	return configurationClass.apiClusterIndex
}

func (configurationClass *ConfigurationClass) GetApiClusterName() string {
	configurationClass.init()
	return configurationClass.openshiftConfig.APICluster
}
