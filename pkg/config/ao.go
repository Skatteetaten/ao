package config

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/prompt"
	"io/ioutil"
	"os"
)

type AOConfig struct {
	APICluster  string              `json:"apiCluster"`
	Affiliation string              `json:"affiliation"`
	Localhost   bool                `json:"localhost"`
	Clusters    map[string]*Cluster `json:"clusters"`

	AvailableClusters       []string `json:"availableClusters"`
	PreferredAPIClusters    []string `json:"preferredApiClusters"`
	AvailableUpdateClusters []string `json:"availableUpdateClusters"`
	ClusterUrlPattern       string   `json:"clusterUrlPattern"`
	BooberUrlPattern        string   `json:"booberUrlPattern"`
	UpdateUrlPattern        string   `json:"updateUrlPattern"`
}

var DefaultAOConfig = AOConfig{
	Clusters:                make(map[string]*Cluster),
	AvailableClusters:       []string{"utv", "utv-relay", "test", "test-relay", "prod", "prod-relay", "qa"},
	PreferredAPIClusters:    []string{"utv", "test"},
	AvailableUpdateClusters: []string{"utv", "test"},
	ClusterUrlPattern:       "https://%s-master.paas.skead.no:8443",
	BooberUrlPattern:        "http://boober-aurora.%s.paas.skead.no",
	UpdateUrlPattern:        "http://ao-aurora-tools.%s.paas.skead.no",
}

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

func WriteConfig(ao AOConfig, configLocation string) error {
	data, err := json.MarshalIndent(ao, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configLocation, data, 0644)
}

func (ao *AOConfig) SelectApiCluster() {
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

func (ao *AOConfig) Update(noPrompt bool) error {
	url := ao.getUpdateUrl()
	if url == "" {
		return errors.New("No update server is available, check config")
	}
	serverVersion, err := GetCurrentVersionFromServer(url)
	if err != nil {
		return err
	}

	if !serverVersion.IsNewVersion() {
		return errors.New("No update available")
	}

	if !noPrompt {
		message := fmt.Sprintf("Do you want update AO from version %s -> %s?", Version, serverVersion.Version)
		update := prompt.Confirm(message, true)
		if !update {
			return errors.New("Update aborted")
		}
	}

	data, err := GetNewAOClient(url)
	if err != nil {
		return err
	}

	return ao.replaceAO(data)
}

func (ao *AOConfig) replaceAO(data []byte) error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	releasePath := executablePath + "_" + "update"
	err = ioutil.WriteFile(releasePath, data, 0750)
	if err != nil {
		return err
	}

	return os.Rename(releasePath, executablePath)
}

func (ao *AOConfig) getUpdateUrl() string {
	var updateCluster string
	for _, c := range ao.AvailableUpdateClusters {
		available, found := ao.Clusters[c]
		logrus.WithField("exists", found).Info("update server", c)
		if found && available.Reachable {
			updateCluster = c
			break
		}
	}

	if updateCluster == "" {
		return ""
	}

	return fmt.Sprintf(ao.UpdateUrlPattern, updateCluster)
}
