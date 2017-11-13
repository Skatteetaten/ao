package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/log"
	"github.com/spf13/cobra"
	"os"
)

// TODO: UPDATE DOCUMENTATION

const rootLong = `A command line interface that interacts with the Boober API
to enable the user to manipulate the Aurora Config for an affiliation, and to
 deploy one or more application.

This application has two main parts.
1. manage the AuroraConfig configuration via cli
2. apply the aoc configuration to the clusters`

var (
	logLevel        string
	prettyLog       bool
	persistentHost  string
	persistentToken string

	// DefaultApiClient will use APICluster from ao config as default values
	// if persistent token and/or server api url is specified these will override default values
	DefaultApiClient *client.ApiClient
	ao               *config.AOConfig
	configLocation   string
)

// TODO: Replace with InitializeOptions
var persistentOptions cmdoptions.CommonCommandOptions

// TODO: Remove all config references
var oldConfig = &configuration.ConfigurationClass{
	PersistentOptions: &persistentOptions,
}

var RootCmd = &cobra.Command{
	Use:               "ao",
	Short:             "Aurora Openshift CLI",
	Long:              rootLong,
	SilenceUsage:      true,
	PersistentPreRunE: Initialize,
	RunE:              ShowAoHelp,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "", "fatal", "Set loglevel. Valid log levels are [info, debug, warning, error, fatal]")
	RootCmd.PersistentFlags().BoolVarP(&prettyLog, "prettylog", "", false, "Pretty print log")
	RootCmd.PersistentFlags().StringVarP(&persistentHost, "serverapi", "", "", "Override default server API address")
	RootCmd.PersistentFlags().StringVarP(&persistentToken, "token", "", "", "Token to be used for serverapi connections")

}

func ShowAoHelp(cmd *cobra.Command, args []string) error {
	cmd.SetHelpTemplate(customHelpTemplate)
	return cmd.Help()
}

func Initialize(cmd *cobra.Command, args []string) error {

	// Errors will be printed from main
	cmd.SilenceErrors = true
	// Disable print usage when an error occurs
	cmd.SilenceUsage = true

	home, _ := os.LookupEnv("HOME")
	configLocation = home + "/.ao.json"

	err := setLogging(logLevel, prettyLog)
	if err != nil {
		return err
	}

	aoConfig, err := config.LoadConfigFile(configLocation)
	if err != nil {
		logrus.Error(err)
	}

	if aoConfig == nil {
		logrus.Info("Creating new config")
		aoConfig = &config.DefaultAOConfig
		aoConfig.InitClusters()
		aoConfig.SelectApiCluster()
		aoConfig.Write(configLocation)
	}

	apiCluster := aoConfig.Clusters[aoConfig.APICluster]
	if apiCluster == nil {
		return errors.Errorf("Api cluster %s is not available. Check config.", aoConfig.APICluster)
	}

	api := client.NewApiClient(apiCluster.BooberUrl, apiCluster.Token, aoConfig.Affiliation)

	if persistentHost != "" {
		api.Host = persistentHost
	} else if aoConfig.Localhost {
		// TODO: Move to config?
		api.Host = "http://localhost:8080"
	}

	if persistentToken != "" {
		api.Token = persistentToken
	}

	ao, DefaultApiClient = aoConfig, api

	return nil
}

func setLogging(level string, pretty bool) error {
	logrus.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	if pretty {
		logrus.SetFormatter(&log.PrettyFormatter{})
	}

	return nil
}

const customHelpTemplate = `{{.Long}}

Usage:
  {{.CommandPath}} [command] [flags]

Basic Commands:{{range .Commands}}{{if (and (eq (index .Annotations "type") "") (ne .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

File Commands:{{range .Commands}}{{if eq (index .Annotations "type") "file"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Util Commands:{{range .Commands}}{{if eq (index .Annotations "type") "util"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
