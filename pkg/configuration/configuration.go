package configuration

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/spf13/viper"
)

type ConfigurationClass struct {
	configLocation    string
	openshiftConfig   *openshift.OpenshiftConfig
	apiClusterIndex   int
	apiClusterName    string
	persistentOptions *cmdoptions.CommonCommandOptions
	initDone          bool
}

func (configuration *ConfigurationClass) Init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	configuration.persistentOptions = persistentOptions
	configuration.init()
	return
}

func (configuration *ConfigurationClass) init() (err error) {
	if configuration.initDone {
		return
	}

	configuration.configLocation = viper.GetString("HOME") + "/.aoc.json"
	configuration.openshiftConfig, err = openshift.LoadOrInitiateConfigFile(configuration.configLocation, false)
	if err != nil {
		err = errors.New("Error in loading OpenShift configuration")
	}
	// Find index for API cluster,that is the first reachable cluster
	if configuration.openshiftConfig != nil {
		for i := range configuration.openshiftConfig.Clusters {
			if configuration.openshiftConfig.Clusters[i].Name == configuration.openshiftConfig.APICluster {
				configuration.apiClusterIndex = i
				break
			}
		}
	}
	configuration.initDone = true
	return
}

func (configuration *ConfigurationClass) GetOpenshiftConfig() *openshift.OpenshiftConfig {
	configuration.init()
	return configuration.openshiftConfig
}

func (configuration *ConfigurationClass) GetApiClusterIndex() int {
	configuration.init()
	return configuration.apiClusterIndex
}

func (configuration *ConfigurationClass) GetApiClusterName() string {
	configuration.init()
	return configuration.openshiftConfig.APICluster
}

func (configuration *ConfigurationClass) GetAffiliation() string {
	configuration.init()
	return configuration.openshiftConfig.Affiliation
}

func (configuration *ConfigurationClass) GetPersistentOptions() *cmdoptions.CommonCommandOptions {
	configuration.init()
	return configuration.persistentOptions
}
