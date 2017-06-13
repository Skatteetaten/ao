package kubernetes

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

func (kubeConfig *KubeConfig) getUserAndToken() (username string, token string) {

	return
}
