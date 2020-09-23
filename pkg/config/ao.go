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
	RefName     string              `json:"refName"`
	APICluster  string              `json:"apiCluster"`
	Affiliation string              `json:"affiliation"`
	Localhost   bool                `json:"localhost"`
	Clusters    map[string]*Cluster `json:"clusters"`

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

// DefaultAOConfig is an AOConfig with default values
var DefaultAOConfig = AOConfig{
	RefName:                 "master",
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
func (ao *AOConfig) GetServiceURLs(clusterName string) (*ServiceURLs, error) {
	if len(ao.ServiceURLPatterns) == 0 {
		return &ServiceURLs{
			BooberURL:       fmt.Sprintf(ao.BooberURLPattern, clusterName),
			ClusterURL:      fmt.Sprintf(ao.ClusterURLPattern, clusterName),
			ClusterLoginURL: fmt.Sprintf(ao.ClusterURLPattern, clusterName),
			GoboURL:         fmt.Sprintf(ao.GoboURLPattern, clusterName),
		}, nil
	}

	clusterConfig := ao.ClusterConfig[clusterName]
	if clusterConfig == nil || clusterConfig.Type == "" {
		return nil, errors.Errorf("missing cluster type for cluster %s", clusterName)
	}

	patterns := ao.ServiceURLPatterns[clusterConfig.Type]
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
func (ao *AOConfig) AddMultipleClusterConfig() {
	newCluster := "utv04"
	ao.AvailableClusters = append(ao.AvailableClusters, newCluster)
	ao.AvailableUpdateClusters = append([]string{newCluster}, ao.AvailableUpdateClusters...)
	ao.ClusterConfig = map[string]*ClusterConfig{
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
		newCluster: {
			Type: "ocp4",
		},
	}
	ao.ServiceURLPatterns = map[string]*ServiceURLPatterns{
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
func WriteConfig(ao AOConfig, configLocation string) error {
	data, err := json.MarshalIndent(ao, "", "  ")
	if err != nil {
		return fmt.Errorf("While marshaling ao config: %w", err)
	}
	if err := ioutil.WriteFile(configLocation, data, 0644); err != nil {
		return fmt.Errorf("While writing ao config to file: %w", err)
	}

	return nil
}

// SelectAPICluster returns specified APICluster or makes a priority based selection of an APICluster
func (ao *AOConfig) SelectAPICluster() {
	if ao.APICluster != "" {
		return
	}

	for _, name := range ao.PreferredAPIClusters {
		cluster, found := ao.Clusters[name]
		if !found {
			continue
		}

		if cluster.Reachable {
			ao.APICluster = name
			return
		}
	}

	for k, cluster := range ao.Clusters {
		if cluster.Reachable {
			ao.APICluster = k
			return
		}
	}
}

// Update checks for a new version of ao and performs update with an optional interactive confirmation
func (ao *AOConfig) Update(noPrompt bool) error {
	url, err := ao.getUpdateURL()
	if err != nil {
		return err
	}

	serverVersion, err := GetCurrentVersionFromServer(url)
	if err != nil {
		return err
	}

	if !serverVersion.IsNewVersion() {
		return errors.New("No update available")
	}

	if !noPrompt {
		if runtime.GOOS == "windows" {
			message := fmt.Sprintf("New version of AO is available (%s) - please download from %s", serverVersion.Version, url)
			fmt.Println(message)
			return nil
		}
		message := fmt.Sprintf("Do you want to update AO from version %s -> %s?", Version, serverVersion.Version)
		update := prompt.Confirm(message, true)
		if !update {
			return errors.New("Update aborted")
		}
	}

	data, err := GetNewAOClient(url)
	if err != nil {
		return err
	}

	err = ao.replaceAO(data)
	if err != nil {
		return err
	}

	return nil
}

func (ao *AOConfig) replaceAO(data []byte) error {
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

func (ao *AOConfig) getUpdateURL() (string, error) {
	for _, cluster := range ao.AvailableUpdateClusters {
		available, found := ao.Clusters[cluster]
		logrus.WithField("exists", found).Info("update server", cluster)

		if !found || (found && !available.Reachable) {
			continue
		}

		updateURL, err := ao.resolveUpdateURLPattern(cluster)
		if err != nil {
			logrus.WithField("cluster", available.Name).Warn(err)
			continue
		}

		return updateURL, nil
	}

	return "", errors.New("could not find any available update servers")
}

func (ao *AOConfig) resolveUpdateURLPattern(clusterName string) (string, error) {
	if len(ao.ServiceURLPatterns) == 0 {
		return fmt.Sprintf(ao.ClusterURLPattern, clusterName), nil
	}

	clusterConfig := ao.ClusterConfig[clusterName]
	if clusterConfig == nil || clusterConfig.Type == "" {
		return "", errors.Errorf("missing cluster type for cluster %s", clusterName)
	}

	patterns := ao.ServiceURLPatterns[clusterConfig.Type]
	if patterns == nil {
		return "", errors.Errorf("missing serviceUrlPatterns for cluster type %s", clusterConfig.Type)
	}

	return formatNonLocalhostPattern(patterns.UpdateURLPattern, clusterName), nil
}
