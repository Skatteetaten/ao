package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

const OCP3 = "ocp3"
const OCP4 = "ocp4"

var ocp3Clusters = []string{"utv", "utv-relay", "test", "test-relay", "prod", "prod-relay"}
var ocp4Clusters = []string{"utv04", "utv05", "utv-relay01", "test01", "test-relay01", "prod01", "prod-relay01", "log01"}
var availableUpdateClusters = []string{"utv", "test", "utv04", "test01"}

// ServiceURLPatterns contains url patterns for all integrations made with AO.
// %s will be replaced with cluster name. If ClusterURLPrefix in ClusterConfig is specified
// it will be used for ClusterURLPattern and ClusterLoginURLPattern insted of cluster name.
type ServiceURLPatterns struct {
	ClusterURLPattern      string `json:"clusterUrlPattern"`
	ClusterLoginURLPattern string `json:"clusterLoginUrlPattern"`
	BooberURLPattern       string `json:"booberUrlPattern"`
	UpdateURLPattern       string `json:"updateUrlPattern"`
	GoboURLPattern         string `json:"goboUrlPattern"`
}

var ocp3URLPatterns = &ServiceURLPatterns{
	ClusterURLPattern:      "https://%s-master.paas.skead.no:8443",
	ClusterLoginURLPattern: "https://%s-master.paas.skead.no:8443",
	BooberURLPattern:       "http://boober-aurora.%s.paas.skead.no",
	UpdateURLPattern:       "http://ao-aurora-tools.%s.paas.skead.no",
	GoboURLPattern:         "http://gobo.aurora.%s.paas.skead.no",
}

var ocp4URLPatterns = &ServiceURLPatterns{
	ClusterURLPattern:      "https://api.%s.paas.skead.no:6443",
	ClusterLoginURLPattern: "https://oauth-openshift.apps.%s.paas.skead.no",
	BooberURLPattern:       "https://boober-aup.apps.%s.paas.skead.no",
	UpdateURLPattern:       "https://ao-aup.apps.%s.paas.skead.no",
	GoboURLPattern:         "https://gobo-aup.apps.%s.paas.skead.no",
}

// AOConfig is a structure of the configured URLs to access clusters used by ao
type AOConfig struct {
	Clusters map[string]*Cluster `json:"clusters"`

	AvailableClusters       []string `json:"availableClusters"` // needed for fixed order, which is not supported by Clusters map
	PreferredAPIClusters    []string `json:"preferredApiClusters"`
	AvailableUpdateClusters []string `json:"availableUpdateClusters"`

	FileAOVersion string `json:"aoVersion"` // For detecting possible changes to saved file
}

func CreateDefaultAoConfig() *AOConfig {
	aoConfig := createMultipleClusterConfig()
	aoConfig.InitClusters()
	return aoConfig
}

func LoadOrCreateAOConfig(customConfigLocation string) *AOConfig {
	customAOConfig := LoadConfigFile(customConfigLocation)

	if customAOConfig == nil {
		logrus.Info("Creating default ao config")
		return CreateDefaultAoConfig()
	}
	if customAOConfig.FileAOVersion != Version {
		fmt.Printf("WARNING: A custom ao config file is saved with another version at %s.\n"+
			"AO-version: %s, saved version: %s\n"+
			"This may cause unforeseen errors.\n", customConfigLocation, Version, customAOConfig.FileAOVersion)
	}
	logrus.Info("Using custom ao config from file")
	return customAOConfig
}

// LoadConfigFile loads an AOConfig file from file system
func LoadConfigFile(configLocation string) *AOConfig {
	raw, err := ioutil.ReadFile(configLocation)
	if err != nil {
		logrus.Debugf("Could not read optional file %s, got: %v "+
			"NB: It is normal that this file does not exist.\n", configLocation, err)
		return nil
	}

	var config *AOConfig
	err = json.Unmarshal(raw, &config)
	if err != nil {
		logrus.Errorf("Could not parse file %s, got error: %v\n", configLocation, err)
		fmt.Printf("WARNING: The custom ao config file at %s did not match expected format.\n"+
			"Consider removing it or generate a new one. Using default ao config.\n", configLocation)
		return nil
	}

	return config
}

// WriteConfig writes an AOConfig file to file system
func WriteConfig(aoConfig AOConfig, configLocation string) error {
	data, err := json.MarshalIndent(aoConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("While marshaling ao config: %w", err)
	}
	if err := ioutil.WriteFile(configLocation, data, 0644); err != nil {
		return fmt.Errorf("While writing ao config to file: %w", err)
	}

	return nil
}

// createMultipleClusterConfig adds a richer cluster configuration for multiple cluster types.
func createMultipleClusterConfig() *AOConfig {
	aoConfig := AOConfig{
		Clusters:                make(map[string]*Cluster),
		AvailableClusters:       append(ocp3Clusters, ocp4Clusters...),
		PreferredAPIClusters:    []string{"utv", "test", "utv04", "test01"},
		AvailableUpdateClusters: availableUpdateClusters,
		FileAOVersion:           Version,
	}

	return &aoConfig
}

// InitClusters initializes Cluster objects for AOConfig
func (aoConfig *AOConfig) InitClusters() {
	aoConfig.Clusters = make(map[string]*Cluster)
	ch := make(chan *Cluster)
	configuredClusters := 0

	// ocp3
	for _, clusterName := range ocp3Clusters {
		cluster := getAoConfigCluster(ocp3URLPatterns, clusterName)
		configuredClusters++
		go checkReachable(ch, &cluster)
	}
	// ocp4
	for _, clusterName := range ocp4Clusters {
		cluster := getAoConfigCluster(ocp4URLPatterns, clusterName)
		configuredClusters++
		go checkReachable(ch, &cluster)
	}

	for {
		select {
		case c := <-ch:
			aoConfig.Clusters[c.Name] = c
			if len(aoConfig.Clusters) == configuredClusters {
				return
			}
		}
	}
}

func (aoConfig *AOConfig) getUpdateURL() (string, error) {
	for _, cluster := range aoConfig.AvailableUpdateClusters {
		available, found := aoConfig.Clusters[cluster]
		logrus.WithField("exists", found).Info("update server", cluster)

		if !found || (found && !available.Reachable) {
			continue
		}
		updateURL := aoConfig.Clusters[cluster].UpdateURL

		return updateURL, nil
	}

	return "", errors.New("could not find any available update servers")
}

// SelectAPICluster returns specified APICluster or makes a priority based selection of an APICluster
func (aoConfig *AOConfig) SelectAPICluster() string {
	for _, name := range aoConfig.PreferredAPIClusters {
		cluster, found := aoConfig.Clusters[name]
		if !found {
			continue
		}

		if cluster.Reachable {
			return name
		}
	}

	for clusterName, cluster := range aoConfig.Clusters {
		if cluster.Reachable {
			return clusterName
		}
	}
	return ""
}

func getAoConfigCluster(serviceUrlPatterns *ServiceURLPatterns, clusterName string) Cluster {
	updateUrl := ""
	if contains(availableUpdateClusters, clusterName) {
		updateUrl = formatNonLocalhostPattern(serviceUrlPatterns.UpdateURLPattern, clusterName)
	}
	return Cluster{
		Name:      clusterName,
		URL:       formatNonLocalhostPattern(serviceUrlPatterns.ClusterURLPattern, clusterName),
		LoginURL:  formatNonLocalhostPattern(serviceUrlPatterns.ClusterLoginURLPattern, clusterName),
		Reachable: false,
		BooberURL: formatNonLocalhostPattern(serviceUrlPatterns.BooberURLPattern, clusterName),
		GoboURL:   formatNonLocalhostPattern(serviceUrlPatterns.GoboURLPattern, clusterName),
		UpdateURL: updateUrl,
	}
}

func checkReachable(ch chan *Cluster, cluster *Cluster) {
	reachable := false
	resp, err := client.Get(cluster.BooberURL)
	if err == nil && resp != nil && resp.StatusCode < 500 {
		resp, err := client.Get(cluster.GoboURL)
		if err == nil && resp != nil && resp.StatusCode < 500 {
			resp, err = client.Get(cluster.LoginURL)
			if err == nil && resp != nil && resp.StatusCode < 500 {
				reachable = true
			}
		}
	}
	cluster.Reachable = reachable
	logrus.WithField("reachable", reachable).Info(cluster.BooberURL)

	ch <- cluster
}

func formatNonLocalhostPattern(pattern string, a ...interface{}) string {
	if strings.Contains(pattern, "localhost") {
		return pattern
	}

	return fmt.Sprintf(pattern, a...)
}

// Checks if a string slice contains a string
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
