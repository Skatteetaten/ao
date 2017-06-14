package kubernetes

import (
	"errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type KubeConfig struct {
	ApiVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			Server string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster   string `yaml:"cluster"`
			Namespace string `yaml:"namespace"`
			User      string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Users          []struct {
		Name string `yaml:"name"`
		User struct {
			Token string `yaml:"token"`
		} `yaml:"user"`
	}
}

func (kubeConfig *KubeConfig) GetConfig() error {

	var kubeConfigLocation = viper.GetString("HOME") + "/.kube/config"

	yamlFile, err := ioutil.ReadFile(kubeConfigLocation)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}

func (kubeConfig *KubeConfig) GetClusterUserAndToken() (clusterAddress string, username string, token string, err error) {

	err = kubeConfig.GetConfig()
	if err != nil {
		return
	}

	currentContext := kubeConfig.CurrentContext
	if currentContext == "" {
		err = errors.New("No current OC context")
		return
	}

	currentContextParts := strings.Split(currentContext, "/")
	if len(currentContextParts) < 3 {
		err = errors.New("Unexpected current context format: " + currentContext)
		return
	}

	currentClusterName := currentContextParts[1]
	username = currentContextParts[2]

	for i := range kubeConfig.Clusters {
		if kubeConfig.Clusters[i].Name == currentClusterName {
			clusterAddress = kubeConfig.Clusters[i].Cluster.Server
		}
	}
	if clusterAddress == "" {
		err = errors.New("Cluster address not found in kubeconfig")
		return
	}

	for i := range kubeConfig.Users {
		if kubeConfig.Users[i].Name == username+"/"+currentClusterName {
			token = kubeConfig.Users[i].User.Token
		}
	}

	if token == "" {
		err = errors.New("Token not found in kubeconfig")
		return
	}

	return
}
