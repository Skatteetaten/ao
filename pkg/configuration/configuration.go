package configuration

import (
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/openshift"
)

type ConfigurationClass struct {
	PersistentOptions *cmdoptions.CommonCommandOptions
	OpenshiftConfig   *openshift.OpenshiftConfig
	apiClusterIndex   int
	apiClusterName    string
	apiClusterUrl     string
	Testing           bool
}

func NewTestConfiguration() (config *ConfigurationClass) {
	config = new(ConfigurationClass)
	config.OpenshiftConfig = new(openshift.OpenshiftConfig)
	config.PersistentOptions = new(cmdoptions.CommonCommandOptions)
	config.Testing = true
	return config
}

func (configuration *ConfigurationClass) SetApiCluster() error {

	// Find index for API cluster,that is the first reachable cluster
	if configuration.OpenshiftConfig != nil {
		for i := range configuration.OpenshiftConfig.Clusters {
			if configuration.OpenshiftConfig.Clusters[i].Name == configuration.OpenshiftConfig.APICluster {
				configuration.apiClusterIndex = i
				configuration.apiClusterUrl = configuration.OpenshiftConfig.Clusters[i].BooberUrl
				break
			}
		}
	}

	return nil
}
func (configuration *ConfigurationClass) GetApiCluster() *openshift.OpenshiftCluster {
	apiCluster := &openshift.OpenshiftCluster{ }
	for _, cluster := range configuration.OpenshiftConfig.Clusters {
		if cluster.Name == configuration.GetApiClusterName() {
			apiCluster = cluster
			break
		}
	}

	return apiCluster
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
