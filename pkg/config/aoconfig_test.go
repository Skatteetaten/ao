package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const configTmpFile = "/tmp/ao-config_test.json"

func TestLoadConfigFile(t *testing.T) {
	defer os.Remove(configTmpFile)
	aoConfig := LoadConfigFile(configTmpFile)
	assert.Empty(t, aoConfig)

	aoConfig = basicTestConfig()
	WriteConfig(*aoConfig, configTmpFile)

	aoConfig = LoadConfigFile(configTmpFile)
	assert.NotEmpty(t, aoConfig)
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
		aoConfig := basicTestConfig()
		aoConfig.Clusters = make(map[string]*Cluster)
		for name, reachable := range test.Clusters {
			aoConfig.Clusters[name] = &Cluster{
				Reachable: reachable,
			}
		}

		apiCluster := aoConfig.SelectAPICluster()
		assert.Equal(t, test.Expected, apiCluster)
		assert.Len(t, aoConfig.Clusters, len(test.Clusters))
	}
}

func basicTestConfig() *AOConfig {
	aoConfig := createMultipleClusterConfig()
	aoConfig.Clusters["test"] = &Cluster{
		Name:      "test",
		URL:       "https://test.url.paas.skead.no:8443",
		LoginURL:  "https://test.login.paas.skead.no:8443",
		Reachable: true,
		BooberURL: "https://boober.test.paas.skead.no",
		GoboURL:   "https://gobo.test.paas.skead.no",
		UpdateURL: "http://ao-update.test.paas.skead.no",
	}
	return aoConfig
}
