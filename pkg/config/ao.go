package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/prompt"
)

var ocp3URLPatterns = &ServiceURLPatterns{
	ClusterURLPattern: "https://%s-master.paas.skead.no:8443",
	BooberURLPattern:  "http://boober-aurora.%s.paas.skead.no",
	UpdateURLPattern:  "http://ao-aurora-tools.%s.paas.skead.no",
	GoboURLPattern:    "http://gobo.aurora.%s.paas.skead.no",
}

var ocp4URLPatterns = &ServiceURLPatterns{
	ClusterURLPattern:      "https://api.%s.paas.skead.no:6443",
	ClusterLoginURLPattern: "https://oauth-openshift.apps.%s.paas.skead.no",
	BooberURLPattern:       "https://boober-aup.apps.%s.paas.skead.no",
	UpdateURLPattern:       "https://ao-aup.apps.%s.paas.skead.no",
	GoboURLPattern:         "https://gobo-aup.apps.%s.paas.skead.no",
}

// ClusterConfig information about features and configuration for a cluster.
type ClusterConfig struct {
	Type             string `json:"type"`
	IsAPICluster     bool   `json:"isApiCluster"`
	IsUpdateCluster  bool   `json:"isUpdateCluster"`
	ClusterURLPrefix string `json:"clusterUrlPrefix"`
}

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

// ServiceURLs contains all the necessary URLs for integrations made with AO.
type ServiceURLs struct {
	BooberURL       string
	ClusterURL      string
	ClusterLoginURL string
	GoboURL         string
}

// AOConfig is a structure of the configuration of ao
type AOConfig struct {
	Clusters map[string]*Cluster `json:"clusters"`

	ServiceURLPatterns map[string]*ServiceURLPatterns `json:"serviceURLPatterns"`
	ClusterConfig      map[string]*ClusterConfig      `json:"clusterConfig"`

	AvailableClusters       []string `json:"availableClusters"`
	PreferredAPIClusters    []string `json:"preferredApiClusters"`
	AvailableUpdateClusters []string `json:"availableUpdateClusters"`
	ClusterURLPattern       string   `json:"clusterUrlPattern"`
	BooberURLPattern        string   `json:"booberUrlPattern"`
	UpdateURLPattern        string   `json:"updateUrlPattern"`
	GoboURLPattern          string   `json:"goboUrlPattern"`

	FileAOVersion string `json:"aoVersion"` // For detecting possible changes to saved file
}

// basicAOConfig is an AOConfig with default values
var basicAOConfig = AOConfig{
	Clusters:                make(map[string]*Cluster),
	AvailableClusters:       []string{"utv", "utv-relay", "test", "test-relay", "prod", "prod-relay"},
	PreferredAPIClusters:    []string{"utv", "test"},
	AvailableUpdateClusters: []string{"utv", "test"},
	ClusterURLPattern:       ocp3URLPatterns.ClusterURLPattern,
	BooberURLPattern:        ocp3URLPatterns.BooberURLPattern,
	UpdateURLPattern:        ocp3URLPatterns.UpdateURLPattern,
	GoboURLPattern:          ocp3URLPatterns.GoboURLPattern,
	FileAOVersion:           Version,
}

// GetServiceURLs returns old config if ServiceURLPatterns is empty, else ServiceURLs for a given cluster type
func (aoConfig *AOConfig) GetServiceURLs(clusterName string) (*ServiceURLs, error) {
	if len(aoConfig.ServiceURLPatterns) == 0 {
		return &ServiceURLs{
			BooberURL:       fmt.Sprintf(aoConfig.BooberURLPattern, clusterName),
			ClusterURL:      fmt.Sprintf(aoConfig.ClusterURLPattern, clusterName),
			ClusterLoginURL: fmt.Sprintf(aoConfig.ClusterURLPattern, clusterName),
			GoboURL:         fmt.Sprintf(aoConfig.GoboURLPattern, clusterName),
		}, nil
	}

	clusterConfig := aoConfig.ClusterConfig[clusterName]
	if clusterConfig == nil || clusterConfig.Type == "" {
		return nil, errors.Errorf("missing cluster type for cluster %s", clusterName)
	}

	patterns := aoConfig.ServiceURLPatterns[clusterConfig.Type]
	if patterns == nil {
		return nil, errors.Errorf("missing serviceUrlPatterns for cluster type %s", clusterConfig.Type)
	}

	clusterPrefix := clusterName
	if clusterConfig.ClusterURLPrefix != "" {
		clusterPrefix = clusterConfig.ClusterURLPrefix
	}

	clusterLoginURLPattern := patterns.ClusterURLPattern
	if patterns.ClusterLoginURLPattern != "" {
		clusterLoginURLPattern = patterns.ClusterLoginURLPattern
	}

	return &ServiceURLs{
		BooberURL:       formatNonLocalhostPattern(patterns.BooberURLPattern, clusterName),
		ClusterURL:      formatNonLocalhostPattern(patterns.ClusterURLPattern, clusterPrefix),
		ClusterLoginURL: formatNonLocalhostPattern(clusterLoginURLPattern, clusterPrefix),
		GoboURL:         formatNonLocalhostPattern(patterns.GoboURLPattern, clusterName),
	}, nil
}

// AddMultipleClusterConfig adds a richer cluster configuration for multiple cluster types.
func (aoConfig *AOConfig) AddMultipleClusterConfig() {
	aoConfig.ClusterConfig = map[string]*ClusterConfig{
		"utv": {
			Type: "ocp3",
		},
		"utv-relay": {
			Type: "ocp3",
		},
		"test": {
			Type: "ocp3",
		},
		"test-relay": {
			Type: "ocp3",
		},
		"prod": {
			Type: "ocp3",
		},
		"prod-relay": {
			Type: "ocp3",
		},
	}
	ocp4Clusters := []string{"utv04", "utv-relay01", "test01", "test-relay01", "prod01", "prod-relay01"}
	for _, cluster := range ocp4Clusters {
		if !contains(aoConfig.AvailableClusters, cluster) {
			aoConfig.AvailableClusters = append(aoConfig.AvailableClusters, cluster)
		}
		aoConfig.ClusterConfig[cluster] = &ClusterConfig{
			Type: "ocp4",
		}
	}
	aoConfig.PreferredAPIClusters = append([]string{"utv04", "test01"}, aoConfig.PreferredAPIClusters...)
	// Oppgraderingsservere for ocp4:
	aoConfig.AvailableUpdateClusters = append([]string{"utv04", "test01"}, aoConfig.AvailableUpdateClusters...)

	aoConfig.ServiceURLPatterns = map[string]*ServiceURLPatterns{
		"ocp3": ocp3URLPatterns,
		"ocp4": ocp4URLPatterns,
	}
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

func LoadOrCreateAOConfig(customConfigLocation string) (*AOConfig, error) {
	customAOConfig, err := LoadConfigFile(customConfigLocation)
	if err != nil {
		// The normal state is that there is no custom AO config file
		logrus.Debug(err)
	}
	if customAOConfig == nil {
		logrus.Info("Creating default config")
		return CreateDefaultConfig(), nil
	}
	if customAOConfig.FileAOVersion != Version {
		logrus.Warnf("A custom ao config file is saved with another version at %s.\n"+
			"AO-version: %s, saved version: %s\n"+
			"This may cause unforeseen errors.\n", customConfigLocation, Version, customAOConfig.FileAOVersion)
	}
	return customAOConfig, nil
}

func CreateDefaultConfig() *AOConfig {
	aoConfig := basicAOConfig
	aoConfig.AddMultipleClusterConfig()
	aoConfig.InitClusters()
	return &aoConfig
}

// LoadConfigFile loads an AOConfig file from file system
func LoadConfigFile(configLocation string) (*AOConfig, error) {
	raw, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return nil, err
	}

	var c *AOConfig
	err = json.Unmarshal(raw, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
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

// Update checks for a new version of ao and performs update with an optional interactive confirmation
// returns true if ao is actually updated
func (aoConfig *AOConfig) Update(noPrompt bool) (bool, error) {
	url, err := aoConfig.getUpdateURL()
	if err != nil {
		return false, err
	}
	logrus.Debugf("Update URL: %s", url)

	serverVersion, err := GetCurrentVersionFromServer(url)
	if err != nil {
		logrus.Warnf("Unable to get ao version from update server on: %s Aborting update detection: %s", url, err)
		return false, nil
	}

	if !serverVersion.IsNewVersion() {
		return false, errors.New("No update available")
	}

	if !noPrompt {
		if runtime.GOOS == "windows" {
			message := fmt.Sprintf("New version of AO is available (%s) - please download from %s", serverVersion.Version, url)
			fmt.Println(message)
			return false, nil
		}
		message := fmt.Sprintf("Do you want to update AO from version %s -> %s?", Version, serverVersion.Version)
		update := prompt.Confirm(message, true)
		if !update {
			return false, errors.New("Update aborted")
		}
	}

	data, err := GetNewAOClient(url)
	if err != nil {
		return false, err
	}

	err = aoConfig.replaceAO(data)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (aoConfig *AOConfig) replaceAO(data []byte) error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	var releasePath string
	// First, we try to write the update to a file in the executable path
	releasePath = executablePath + "_" + "update"
	err = ioutil.WriteFile(releasePath, data, 0755)
	if err != nil {
		// Could not write to executable path, typically because binary is installed in /usr/bin or /usr/local/bin
		// Try the OS Temp Dir
		releasePath = filepath.Join(os.TempDir(), "ao_update")
		err = ioutil.WriteFile(releasePath, data, 0755)
		if err != nil {
			return err
		}
	}
	err = os.Rename(releasePath, executablePath)
	if err != nil {
		err = errors.New("Could not update AO because it is installed in a different file system than temp: " + err.Error())
		return err
	}
	return nil
}

func (aoConfig *AOConfig) getUpdateURL() (string, error) {
	for _, cluster := range aoConfig.AvailableUpdateClusters {
		available, found := aoConfig.Clusters[cluster]
		logrus.WithField("exists", found).Info("update server", cluster)

		if !found || (found && !available.Reachable) {
			continue
		}

		updateURL, err := aoConfig.resolveUpdateURLPattern(cluster)
		if err != nil {
			logrus.WithField("cluster", available.Name).Warn(err)
			continue
		}

		return updateURL, nil
	}

	return "", errors.New("could not find any available update servers")
}

func (aoConfig *AOConfig) resolveUpdateURLPattern(clusterName string) (string, error) {
	if len(aoConfig.ServiceURLPatterns) == 0 {
		return fmt.Sprintf(aoConfig.UpdateURLPattern, clusterName), nil
	}

	clusterConfig := aoConfig.ClusterConfig[clusterName]
	if clusterConfig == nil || clusterConfig.Type == "" {
		return "", errors.Errorf("missing cluster type for cluster %s", clusterName)
	}

	patterns := aoConfig.ServiceURLPatterns[clusterConfig.Type]
	if patterns == nil {
		return "", errors.Errorf("missing serviceUrlPatterns for cluster type %s", clusterConfig.Type)
	}

	return formatNonLocalhostPattern(patterns.UpdateURLPattern, clusterName), nil
}
