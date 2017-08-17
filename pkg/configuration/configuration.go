package configuration

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/spf13/viper"
)

type ConfigurationClass struct {
	PersistentOptions *cmdoptions.CommonCommandOptions
	OpenshiftConfig   *openshift.OpenshiftConfig
	configLocation    string
	apiClusterIndex   int
	apiClusterName    string
	initDone          bool
}

func (configuration *ConfigurationClass) Init() (err error) {
	if configuration.initDone {
		return
	}

	configuration.configLocation = viper.GetString("HOME") + "/.ao.json"
	configuration.OpenshiftConfig, err = openshift.LoadOrInitiateConfigFile(configuration.configLocation, false)
	if err != nil {
		err = errors.New("Error in loading OpenShift configuration")
	}
	// Find index for API cluster,that is the first reachable cluster
	if configuration.OpenshiftConfig != nil {
		for i := range configuration.OpenshiftConfig.Clusters {
			if configuration.OpenshiftConfig.Clusters[i].Name == configuration.OpenshiftConfig.APICluster {
				configuration.apiClusterIndex = i
				break
			}
		}
	}
	configuration.initDone = true
	return
}

func (configuration *ConfigurationClass) GetApiClusterIndex() int {
	return configuration.apiClusterIndex
}

func (configuration *ConfigurationClass) GetApiClusterName() string {
	return configuration.OpenshiftConfig.APICluster
}

func (configuration *ConfigurationClass) GetAffiliation() string {
	return configuration.OpenshiftConfig.Affiliation
}

func (configuration *ConfigurationClass) GetPersistentOptions() *cmdoptions.CommonCommandOptions {
	return configuration.PersistentOptions
}
