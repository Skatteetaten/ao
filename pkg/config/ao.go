package config

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/printutil"
	"io/ioutil"
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

	CheckoutPaths map[string]string `json:"checkoutPaths"`
}

var DefaultAOConfig = AOConfig{
	Clusters:                make(map[string]*Cluster),
	CheckoutPaths:           make(map[string]string),
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

func (ao *AOConfig) Write(configLocation string) error {
	data, err := json.MarshalIndent(ao, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configLocation, data, 0644)
	if err != nil {
		return err
	}
	return nil
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

// TODO: Remove checkout paths?
func (ao *AOConfig) AddCheckoutPath(affiliation, path, configLocation string) error {
	if ao.CheckoutPaths == nil {
		ao.CheckoutPaths = make(map[string]string)
	}
	ao.CheckoutPaths[affiliation] = path
	return ao.Write(configLocation)
}

func (ao *AOConfig) RemoveCheckoutPath(affiliation string, configLocation string) error {
	if ao.CheckoutPaths == nil {
		return errors.New("There are no checkout path to remove")
	}
	delete(ao.CheckoutPaths, affiliation)
	return ao.Write(configLocation)
}

// TODO: Move
func (ao *AOConfig) PrintClusters(clusterName string, allClusters bool) {

	var headers []string = []string{"CLUSTER NAME", "REACHABLE", "LOGGED IN", "API", "URL"}
	var clusterNames []string
	var reachable []string
	var loggedIn []string
	var api []string
	var url []string

	for name, cluster := range ao.Clusters {
		if cluster.Reachable || allClusters {
			if name == clusterName || clusterName == "" {
				clusterNames = append(clusterNames, name)
				reachableColumn := ""
				if cluster.Reachable {
					reachableColumn = "Yes"
				}
				reachable = append(reachable, reachableColumn)

				loggedInColumn := ""
				if cluster.HasValidToken() {
					loggedInColumn = "Yes"
				}
				loggedIn = append(loggedIn, loggedInColumn)

				apiColumn := ""
				if name == ao.APICluster {
					apiColumn = "Yes"
				}
				api = append(api, apiColumn)

				url = append(url, cluster.Url)
			}
		}
	}

	fmt.Println(printutil.FormatTable(headers, clusterNames, reachable, loggedIn, api, url))
}
