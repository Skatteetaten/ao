// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>
package openshift

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/skatteetaten/aoc/pkg/kubernetes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	Name      string `json:"name"`
	Url       string `json:"url"`
	Token     string `json:"token"`
	Reachable bool   `json:"reachable"`
}

type OpenshiftConfig struct {
	APICluster  string              `json:"apiCluster"`
	Affiliation string              `json:"affiliation"`
	Clusters    []*OpenshiftCluster `json:"clusters"`
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

func Login(configLocation string, userName string, affiliation string) {

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
	config.write(configLocation)
}

func LoadOrInitiateConfigFile(configLocation string, useOcConfig bool) (*OpenshiftConfig, error) {
	config, err := loadConfigFile(configLocation)

	if err != nil {
		//fmt.Println("No config file found, initializing new config")
		config, err := newConfig(useOcConfig)
		if err != nil {
			return nil, err
		}
		if err := config.write(configLocation); err != nil {
			return nil, err
		}
		return config, nil
	}
	return config, nil
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

func newConfig(useOcConfig bool) (config *OpenshiftConfig, err error) {
	//fmt.Println("Pinging all clusters and noting which clusters are active in this profile")

	var taxNorwayClusterFound = false
	if !useOcConfig {
		ch := make(chan *OpenshiftCluster)
		clusters := []string{"utv", "test", "prod", "utv-relay", "test-relay", "prod-relay"}
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
			fmt.Println("Please update ~/.aoc.json manually")
			return
		}
		fmt.Println("OC config detected; default cluster config created")
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
			if (currentOcCluster == "" && apiCluster == "" && c.Reachable && !strings.Contains(c.Name, "-relay")) || (c.Name == currentOcCluster && c.Reachable) {
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
