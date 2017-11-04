package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	aoConfig "github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

// TODO: UPDATE DOCUMENTATION

// DefaultApiClient will use APICluster from ao config as default values
// if persistent token and/or server api url is specified these will override default values
var DefaultApiClient *client.ApiClient

var configLocation string

// TODO: rename import aoConfig to config
var ao *aoConfig.AOConfig

// TODO: Change class name
var persistentOptions cmdoptions.CommonCommandOptions

// TODO: Remove all config references
var config = &configuration.ConfigurationClass{
	PersistentOptions: &persistentOptions,
}

var RootCmd = &cobra.Command{
	Use:   "ao",
	Short: "Aurora Openshift CLI",
	Long: `A command line interface that interacts with the Boober API
to enable the user to manipulate the Aurora Config for an affiliation, and to
 deploy one or more application.

This application has two main parts.
1. manage the AuroraConfig configuration via cli
2. apply the aoc configuration to the clusters
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		level, err := logrus.ParseLevel(persistentOptions.LogLevel)
		if err == nil {
			logrus.SetLevel(level)
		} else {
			fmt.Println(err)
		}

		if persistentOptions.Pretty {
			logrus.SetFormatter(&client.PrettyFormatter{})
		}

		apiCluster := ao.Clusters[ao.APICluster]
		if apiCluster == nil {
			apiCluster = &aoConfig.Cluster{}
		}

		DefaultApiClient = client.NewApiClient(apiCluster.BooberUrl, apiCluster.Token, ao.Affiliation)

		if persistentOptions.ServerApi != "" {
			DefaultApiClient.Host = persistentOptions.ServerApi
		}

		if persistentOptions.Token != "" {
			DefaultApiClient.Token = persistentOptions.Token
			// If token flag is specified, ignore login check
			return
		}

		commandsWithoutLogin := []string{"login", "logout", "version", "help", "adm"}

		commands := strings.Split(cmd.CommandPath(), " ")
		if len(commands) > 1 {
			for _, command := range commandsWithoutLogin {
				if commands[1] == command {
					return
				}
			}
		}

		// TODO: Rework this
		if ao.Affiliation == "" && cmd.Name() != "deploy" {
			ao.Affiliation = prompt.Affiliation("Choose")
		}

		user, _ := os.LookupEnv("USER")
		ao.Login(configLocation, aoConfig.LoginOptions{
			UserName: user,
		})

		// Affiliation and api cluster may be changed
		DefaultApiClient.Affiliation = ao.Affiliation
		apiCluster = ao.Clusters[ao.APICluster]
		if DefaultApiClient.Token == "" && apiCluster != nil {
			DefaultApiClient.Token = apiCluster.Token
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	logrus.SetOutput(os.Stdout)

	home, _ := os.LookupEnv("HOME")
	configLocation = home + "/.ao.json"
	conf, err := aoConfig.LoadConfigFile(configLocation)
	if err != nil {
		fmt.Println(err)
	}

	if conf == nil || recreateConfig {
		logrus.Info("Creating new config")
		conf = &aoConfig.DefaultAOConfig
		conf.InitClusters()
		conf.SelectApiCluster()
		conf.Write(configLocation)
	}
	// Set global config variable
	ao = conf

	// TODO: Mark as hidden?
	RootCmd.PersistentFlags().StringVarP(&persistentOptions.LogLevel, "loglevel", "", "fatal", "Set loglevel. Valid log levels are [info, debug, warning, error, fatal]")

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Pretty, "prettylog",
		"", false, "Pretty print log")

	RootCmd.PersistentFlags().StringVarP(&persistentOptions.ServerApi, "serverapi",
		"", "", "Override default server API address")
	RootCmd.PersistentFlags().StringVarP(&persistentOptions.Token, "token",
		"", "", "Token to be used for serverapi connections")

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Localhost, "localhost", "l", false, "Send all request to localhost api on port 8080")
	RootCmd.PersistentFlags().MarkHidden("localhost")
}
