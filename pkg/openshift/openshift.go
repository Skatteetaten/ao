// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>
package openshift

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/howeyc/gopass"
	"github.com/skatteetaten/ao/pkg/kubernetes"
)

const (
	urlPattern = "https://%s-master.paas.skead.no:8443"
)

var (
	transport = http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = http.Client{
		Transport: &transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

type OpenshiftCluster struct {
	Name       string `json:"name"`
	Url        string `json:"url"`
	Token      string `json:"token"`
	Reachable  bool   `json:"reachable"`
	BooberUrl  string `json:"booberUrl"`
	ConsoleUrl string `json:"consoleUrl"`
}

type OpenshiftConfig struct {
	APICluster     string              `json:"apiCluster"`
	Affiliation    string              `json:"affiliation"`
	Clusters       []*OpenshiftCluster `json:"clusters"`
	CheckoutPaths  map[string]string   `json:"checkoutPaths"`
	Localhost      bool                `json:"localhost"`
	ConfigLocation string
}

func getConsoleUrl(clusterName string) (consoleAddress string) {
	consoleAddress = "http://console-aurora." + clusterName + ".paas.skead.no"
	//consoleAddress = "http://console-paas-espen-dev." + clusterName + ".paas.skead.no"
	return consoleAddress
}

func getApiUrl(clusterName string, localhost bool) (apiAddress string) {
	const localhostAddress = "localhost"
	const localhostPort = "8080"

	if localhost {
		apiAddress = "http://" + localhostAddress + ":" + localhostPort
	} else {
		apiAddress = "http://boober-aurora." + clusterName + ".paas.skead.no"
	}
	return apiAddress
}

func Logout(configLocation string) (err error) {
	config, err := loadConfigFile(configLocation)
	if err != nil {
		return
	}

	for idx := range config.Clusters {
		config.Clusters[idx].Token = ""
	}

	config.Affiliation = ""
	err = config.write(configLocation)
	if err != nil {
		return
	}

	return
}

func Login(configLocation string, userName string, affiliation string, apiCluster string, localhost bool, loginCluster string) {

	//fmt.Println("Login in to all reachable cluster with userName", userName)
	config, err := loadConfigFile(configLocation)

	if err != nil {
		log.Fatal(err)
	}

	config.Affiliation = affiliation
	var password string
	for idx := range config.Clusters {
		cluster := config.Clusters[idx]
		if !cluster.Reachable {
			continue
		}
		if cluster.HasValidToken() {
			//fmt.Println("Cluster ", cluster.Name, " has a valid token")
			continue
		}
		if loginCluster != "" && cluster.Name != loginCluster {
			// User will limit login to the given cluster
			continue
		}
		if password == "" {
			pass, err := askForPassword()
			if err != nil {
				log.Fatal(err)
			}
			password = pass
		}
		token, err := getToken(cluster.Url, userName, password)
		if err != nil {
			log.Fatal(err)
		}
		cluster.Token = token
	}
	// IF apiCluster is supplied on the login command, then set the API cluster if it is defined and reachable.
	if apiCluster != "" {
		for idx := range config.Clusters {
			if config.Clusters[idx].Name == apiCluster && config.Clusters[idx].Reachable {
				config.APICluster = apiCluster
			}
		}
	}
	// IF localhost specified, then put that into the config
	config.Localhost = localhost

	// Write the config.
	config.write(configLocation)
}

func LoadOrInitiateConfigFile(configLocation, loginCluster string, useOcConfig bool) (*OpenshiftConfig, error) {
	config, err := loadConfigFile(configLocation)

	var booberUrlFound bool
	if config != nil {
		for i := range config.Clusters {
			if config.Clusters[i].BooberUrl != "" {
				booberUrlFound = true
			}
		}
	}

	if err != nil || !booberUrlFound {
		return CreateAoConfig(configLocation, loginCluster, useOcConfig)
	}

	return config, nil
}

func CreateAoConfig(configLocation, loginCluster string, useOcConfig bool) (*OpenshiftConfig, error) {
	config, err := newConfig(useOcConfig, loginCluster)
	if err != nil {
		return nil, err
	}
	if err := config.write(configLocation); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *OpenshiftConfig) AddCheckoutPath(affiliation string, path string) error {
	if c.CheckoutPaths == nil {
		c.CheckoutPaths = make(map[string]string)
	}

	c.CheckoutPaths[affiliation] = path

	return c.write(c.ConfigLocation)
}

func (c *OpenshiftConfig) RemoveCheckoutPath(affiliation string, configLocation string) error {
	if c.CheckoutPaths == nil {
		return errors.New("There are no checkout path to remove")
	}

	delete(c.CheckoutPaths, affiliation)

	return c.write(configLocation)
}

func loadConfigFile(configLocation string) (*OpenshiftConfig, error) {
	raw, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return nil, err
	}

	var config *OpenshiftConfig
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}
	return config, nil

}

func (openshiftConfig *OpenshiftConfig) GetApiCluster() (cluster *OpenshiftCluster, err error) {
	if openshiftConfig != nil {
		for clusterIndex := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[clusterIndex].Name == openshiftConfig.APICluster {
				cluster = openshiftConfig.Clusters[clusterIndex]
				return
			}
		}
	}
	err = errors.New("No API cluster defined")
	return
}

func (this *OpenshiftCluster) HasValidToken() bool {
	if this.Token == "" {
		return false
	}

	clusterUrl := fmt.Sprintf("%s/%s", this.Url, "oapi")

	resp, err := getBearer(clusterUrl, this.Token)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true

}

func (this *OpenshiftConfig) write(configLocation string) error {
	configJson, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configLocation, configJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Generates config based upon searching for OpenShift nodes
func newConfig(useOcConfig bool, loginCluster string) (config *OpenshiftConfig, err error) {
	//fmt.Println("Pinging all clusters and noting which clusters are active in this profile")

	var taxNorwayClusterFound = false
	if !useOcConfig {
		ch := make(chan *OpenshiftCluster)
		var clusters []string
		if loginCluster != "" {
			clusters = []string{loginCluster}
		} else {
			clusters = []string{"utv", "test", "prod", "utv-relay", "test-relay", "prod-relay", "qa"}
		}

		for _, c := range clusters {
			cluster := fmt.Sprintf(urlPattern, c)
			go newOpenshiftCluster(c, cluster, ch)
		}

		config = collectOpenshiftClusters(len(clusters), ch, "")
		if config != nil {
			for i := range config.Clusters {
				if config.Clusters[i].Reachable {
					taxNorwayClusterFound = true
				}
			}

		}
	}

	if taxNorwayClusterFound {
		fmt.Println("Running in a Norwegian Tax Compliant environment; default cluster config created")

	} else {
		config, err = getOcClusters()
		if err != nil || config == nil {
			config = emptyConfig()
			err = nil
			fmt.Println("No config detected; empty cluster config created.")
			fmt.Println("Please update ~/.ao.json manually")
			return
		}
		fmt.Println("OC config detected; default cluster config created")
	}

	if config != nil {
		for i := range config.Clusters {
			config.Clusters[i].BooberUrl = getApiUrl(config.Clusters[i].Name, false)
			config.Clusters[i].ConsoleUrl = getConsoleUrl(config.Clusters[i].Name)
		}
	}
	return
}

func emptyConfig() (config *OpenshiftConfig) {
	var emptyConfig OpenshiftConfig

	return &emptyConfig
}

func getOcClusters() (config *OpenshiftConfig, err error) {
	var kubeConfig kubernetes.KubeConfig

	err = kubeConfig.GetConfig()
	if err != nil {
		return
	}
	currentOcCluster, err := kubeConfig.GetClusterName()
	if err != nil {
		currentOcCluster = ""
	}

	ch := make(chan *OpenshiftCluster)
	for i := range kubeConfig.Clusters {
		go newOpenshiftCluster(kubeConfig.Clusters[i].Name, kubeConfig.Clusters[i].Cluster.Server, ch)
	}

	config = collectOpenshiftClusters(len(kubeConfig.Clusters), ch, currentOcCluster)

	for i := range config.Clusters {
		var token string
		token, err = kubeConfig.GetToken(config.Clusters[i].Name)
		if err != nil {
			return
		}
		config.Clusters[i].Token = token
	}
	return
}

func newOpenshiftCluster(name string, cluster string, ch chan *OpenshiftCluster) {
	//cluster := fmt.Sprintf(urlPattern, name)
	reachable := ping(cluster)
	ch <- &OpenshiftCluster{
		Name:      name,
		Url:       cluster,
		Reachable: reachable,
	}
}

func collectOpenshiftClusters(num int, ch chan *OpenshiftCluster, currentOcCluster string) *OpenshiftConfig {
	var apiCluster string
	openshiftClusters := []*OpenshiftCluster{}
	for {
		select {
		case c := <-ch:
			openshiftClusters = append(openshiftClusters, c)
			if c.Reachable {
				fmt.Println(c.Name, " is reachable")
			}
			if (currentOcCluster == "" && apiCluster == "" && c.Reachable && !strings.Contains(c.Name, "qa") && !strings.Contains(c.Name, "-relay")) || (c.Name == currentOcCluster && c.Reachable) {
				fmt.Println(c.Name, " is the BooberAPI that will be used")
				apiCluster = c.Name
			}
			if len(openshiftClusters) == num {
				config := &OpenshiftConfig{
					Clusters:   openshiftClusters,
					APICluster: apiCluster,
				}

				return config
			}

		}
	}
}

var timeout = time.Duration(1 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func ping(url string) bool {

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func getBearer(url string, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	tokenValue := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", tokenValue)
	return client.Do(req)
}

func getBasicAuth(url string, username string, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	return client.Do(req)
}

func askForPassword() (string, error) {
	fmt.Printf("Password: ")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		return "", err
	}
	password := string(pass[:])
	return password, nil
}

func getToken(cluster string, username string, password string) (string, error) {
	urlSuffix := "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"
	clusterUrl := cluster + urlSuffix
	resp, err := getBasicAuth(clusterUrl, username, password)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("Not authorized")
	}
	redirectUrl := resp.Header.Get("Location")
	token, err := oauthAuthorizeResult(redirectUrl)
	if err != nil {
		return "", err
	}
	return token, err
}

func oauthAuthorizeResult(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	if errorCode := u.Query().Get("error"); len(errorCode) > 0 {
		errorDescription := u.Query().Get("error_description")
		return "", errors.New(errorCode + " " + errorDescription)
	}

	fragmentValues, err := url.ParseQuery(u.Fragment)
	if err != nil {
		return "", err
	}
	accessToken := fragmentValues.Get("access_token")
	if len(accessToken) == 0 {
		return "", errors.New("token is empty")
	}
	return accessToken, nil
}
