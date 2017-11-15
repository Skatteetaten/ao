package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const configTmpFile = "/tmp/ao_test.json"

func TestLoadConfigFile(t *testing.T) {
	defer os.Remove(configTmpFile)
	ao, _ := LoadConfigFile(configTmpFile)
	assert.Empty(t, ao)

	ao = &DefaultAOConfig

	assert.Empty(t, ao.Affiliation)
	ao.Affiliation = "paas"
	WriteConfig(*ao, configTmpFile)

	ao, _ = LoadConfigFile(configTmpFile)
	assert.NotEmpty(t, ao)

	assert.Equal(t, "paas", ao.Affiliation)
}

func TestAOConfig_SelectApiCluster(t *testing.T) {
	tests := []struct {
		Clusters map[string]bool
		Expected string
	}{
		{map[string]bool{"prod": true, "utv": true, "test": true, "qa": true}, "utv"},
		{map[string]bool{"utv": false, "test": true, "qa": true}, "test"},
		{map[string]bool{"qa": true, "test": false, "utv": false}, "qa"},
	}

	for _, test := range tests {
		aoConfig := DefaultAOConfig
		aoConfig.Clusters = make(map[string]*Cluster)
		for name, reachable := range test.Clusters {
			aoConfig.Clusters[name] = &Cluster{
				Reachable: reachable,
			}
		}

		aoConfig.SelectApiCluster()
		assert.Equal(t, test.Expected, aoConfig.APICluster)
		assert.Len(t, aoConfig.Clusters, len(test.Clusters))
	}

	aoConfig := DefaultAOConfig
	aoConfig.APICluster = "test"
	aoConfig.Clusters["utv"] = &Cluster{
		Reachable: true,
	}

	aoConfig.SelectApiCluster()
	assert.Equal(t, "test", aoConfig.APICluster, "Should not override APICluster when set")
}
