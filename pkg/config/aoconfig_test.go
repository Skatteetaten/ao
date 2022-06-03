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
		{map[string]bool{"prod01": true, "utv04": true, "test01": true, "utv05": true}, "utv04"},
		{map[string]bool{"utv04": false, "test01": true}, "test01"},
		{map[string]bool{"test01": false, "utv05": false, "test": true}, "test"},
		{map[string]bool{"test02": false, "utv01": true}, "utv01"},
		{map[string]bool{"test02": false, "utv04": false, "utv01": false}, ""},
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
	aoConfig.Clusters["test01"] = &Cluster{
		Name:      "test01",
		URL:       "https://api.test01.paas.skead.no:6443",
		LoginURL:  "https://oauth-openshift.apps.test01.paas.skead.no",
		Reachable: true,
		BooberURL: "https://boober-aup.apps.test01.paas.skead.no",
		GoboURL:   "https://gobo-aup.apps.test01.paas.skead.no",
		UpdateURL: "https://ao-aup.apps.test01.paas.skead.no",
	}
	return aoConfig
}
