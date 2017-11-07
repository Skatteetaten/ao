package config

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/prompt"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	transport = http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			timeout := time.Duration(1 * time.Second)
			return net.DialTimeout(network, addr, timeout)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = http.Client{
		Transport: &transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

type Cluster struct {
	Name      string `json:"name"`
	Url       string `json:"url"`
	Token     string `json:"token"`
	Reachable bool   `json:"reachable"`
	BooberUrl string `json:"booberUrl"`
}

func (c *Cluster) HasValidToken() bool {
	if c.Token == "" {
		return false
	}

	clusterUrl := fmt.Sprintf("%s/%s", c.Url, "oapi")
	req, err := http.NewRequest("GET", clusterUrl, nil)
	if err != nil {
		return false
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)
	logrus.Debug("Check for valid token ", clusterUrl)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (ao *AOConfig) InitClusters() {
	ao.Clusters = make(map[string]*Cluster)
	ch := make(chan *Cluster)

	for _, cluster := range ao.AvailableClusters {
		name := cluster
		clusterUrl := fmt.Sprintf(ao.ClusterUrlPattern, name)
		go func() {
			reachable := true
			resp, err := client.Get(clusterUrl)
			if err != nil || resp == nil {
				reachable = false
			}
			if resp != nil && resp.StatusCode != http.StatusOK {
				reachable = false
			}
			logrus.WithField("reachable", reachable).Info(clusterUrl)
			ch <- &Cluster{
				Name:      name,
				Url:       clusterUrl,
				Reachable: reachable,
				BooberUrl: fmt.Sprintf(ao.BooberUrlPattern, name),
			}
		}()
	}

	for {
		select {
		case c := <-ch:
			ao.Clusters[c.Name] = c
			if len(ao.Clusters) == len(ao.AvailableClusters) {
				return
			}
		}
	}
}

func (ao *AOConfig) Logout(configLocation string) error {
	ao.Affiliation = ""
	for _, c := range ao.Clusters {
		c.Token = ""
	}

	ao.Localhost = false

	err := ao.Write(configLocation)
	if err != nil {
		return err
	}

	return nil
}

type LoginOptions struct {
	Affiliation string
	UserName    string
	APICluster  string
	LocalHost   bool
}

// TODO: If localhost bypass cluster check
// TODO: Return error when login fail
func (ao *AOConfig) Login(configLocation string, options LoginOptions) {

	if options.Affiliation != "" {
		ao.Affiliation = options.Affiliation
	}

	if ao.Affiliation == "" {
		ao.Affiliation = prompt.Affiliation("Login")
	}

	var password string
	for _, c := range ao.Clusters {
		if !c.Reachable {
			continue
		}
		if c.HasValidToken() {
			continue
		}
		if password == "" {
			password = prompt.Password()
		}
		token, err := getToken(c.Url, options.UserName, password)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"url":      c.Url,
				"userName": options.UserName,
			}).Fatal(err)
		}
		c.Token = token
	}
	if options.APICluster != "" {
		if cluster, found := ao.Clusters[options.APICluster]; found && cluster.Reachable {
			ao.APICluster = options.APICluster
		} else {
			ao.SelectApiCluster()
			fmt.Printf("Specified api cluster %s is not available, using %s\n", options.APICluster, ao.APICluster)
		}
	}

	ao.Localhost = options.LocalHost
	ao.Write(configLocation)
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

func getBasicAuth(url string, username string, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	return client.Do(req)
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
