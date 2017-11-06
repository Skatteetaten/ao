package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"os"
)

type InitializeOptions struct {
	Host        string
	Token       string
	LogLevel    string
	PrettyLog   bool
	CommandName string
}

func Initialize(configLocation string, options InitializeOptions) (*config.AOConfig, *client.ApiClient) {

	setLogging(options.LogLevel, options.PrettyLog)

	ao, err := config.LoadConfigFile(configLocation)
	if err != nil {
		logrus.Error(err)
	}

	if ao == nil {
		logrus.Info("Creating new config")
		ao = &config.DefaultAOConfig
		ao.InitClusters()
		ao.SelectApiCluster()
		ao.Write(configLocation)
	}

	apiCluster := ao.Clusters[ao.APICluster]
	if apiCluster == nil {
		apiCluster = &config.Cluster{}
	}

	defaultClient := client.NewApiClient(apiCluster.BooberUrl, apiCluster.Token, ao.Affiliation)

	if options.Host != "" {
		defaultClient.Host = options.Host
	} else if ao.Localhost {
		// TODO: Move to config?
		defaultClient.Host = "http://localhost:8080"
	}

	if options.Token != "" {
		defaultClient.Token = options.Token
		// If token flag is specified, ignore login check
		return ao, defaultClient
	}

	commandsWithoutLogin := []string{"login", "logout", "version", "help", "adm"}
	for _, command := range commandsWithoutLogin {
		if options.CommandName == command {
			return ao, defaultClient
		}
	}

	user, _ := os.LookupEnv("USER")
	ao.Login(configLocation, config.LoginOptions{
		UserName: user,
	})

	// TODO: Rework this
	if ao.Affiliation == "" && options.CommandName != "deploy" {
		ao.Affiliation = prompt.Affiliation("Choose")
		defaultClient.Affiliation = ao.Affiliation
	}

	apiCluster = ao.Clusters[ao.APICluster]
	if defaultClient.Token == "" && apiCluster != nil {
		defaultClient.Token = apiCluster.Token
	}

	return ao, defaultClient
}

func setLogging(level string, pretty bool) {
	logrus.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(level)
	if err == nil {
		logrus.SetLevel(lvl)
	} else {
		fmt.Println(err)
	}

	if pretty {
		logrus.SetFormatter(&client.PrettyFormatter{})
	}
}
