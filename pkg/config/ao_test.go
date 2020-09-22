package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

		aoConfig.SelectAPICluster()
		assert.Equal(t, test.Expected, aoConfig.APICluster)
		assert.Len(t, aoConfig.Clusters, len(test.Clusters))
	}

	aoConfig := DefaultAOConfig
	aoConfig.APICluster = "test"
	aoConfig.Clusters["utv"] = &Cluster{
		Reachable: true,
	}

	aoConfig.SelectAPICluster()
	assert.Equal(t, "test", aoConfig.APICluster, "Should not override APICluster when set")
}

func TestAOConfig_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	aoConfig := &AOConfig{
		ClusterURLPattern:       "%s",
		UpdateURLPattern:        "%s/update",
		BooberURLPattern:        "%s",
		GoboURLPattern:          "%s",
		AvailableClusters:       []string{ts.URL},
		AvailableUpdateClusters: []string{ts.URL},
	}

	aoConfig.InitClusters()
	aoConfig.SelectAPICluster()

	url, err := aoConfig.getUpdateURL()

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s/update", ts.URL), url)
}

func TestAOConfig_UpdateWithBetaConfig(t *testing.T) {
	ocp3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ocp4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer ocp3.Close()
	defer ocp4.Close()

	ao := &AOConfig{
		ClusterURLPattern: "%s/old",
		UpdateURLPattern:  "%s/old-update",
		BooberURLPattern:  "%s/old",
		GoboURLPattern:    "%s/old",
		ClusterConfig: map[string]*ClusterConfig{
			ocp3.URL: {
				Type: "ocp3",
			},
			ocp4.URL: {
				Type: "ocp4",
			},
		},
		ServiceURLPatterns: map[string]*ServiceURLPatterns{
			"ocp3": {
				ClusterURLPattern:      "%s/ocp3",
				UpdateURLPattern:       "%s/ocp3-update",
				BooberURLPattern:       "%s/ocp3",
				GoboURLPattern:         "%s/ocp3",
				ClusterLoginURLPattern: "%s/ocp3",
			},
			"ocp4": {
				ClusterURLPattern:      "%s/ocp4",
				UpdateURLPattern:       "%s/ocp4-update",
				BooberURLPattern:       "%s/ocp4",
				GoboURLPattern:         "%s/ocp4",
				ClusterLoginURLPattern: "%s/ocp4",
			},
		},
		AvailableClusters:       []string{ocp4.URL, ocp3.URL},
		AvailableUpdateClusters: []string{ocp4.URL, ocp3.URL},
	}

	// Making both test servers (ocp3, ocp4) reachable
	ao.InitClusters()

	// Should get update URL for ocp4 test server
	url, err := ao.getUpdateURL()

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s/ocp4-update", ocp4.URL), url)
}
