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
	pFlagLogLevel  string
	pFlagPrettyLog bool
	pFlagHost      string
	pFlagToken     string

	// DefaultApiClient will use APICluster from ao config as default values
	// if persistent token and/or server api url is specified these will override default values
	DefaultApiClient *client.ApiClient
	AO               *config.AOConfig
	ConfigLocation   string
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
	PersistentPreRunE: initialize,
	RunE:              showAoHelp,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&pFlagLogLevel, "loglevel", "", "fatal", "Set loglevel. Valid log levels are [info, debug, warning, error, fatal]")
	RootCmd.PersistentFlags().BoolVarP(&pFlagPrettyLog, "prettylog", "", false, "Pretty print log")
	RootCmd.PersistentFlags().StringVarP(&pFlagHost, "serverapi", "", "", "Override default server API address")
	RootCmd.PersistentFlags().StringVarP(&pFlagToken, "token", "", "", "Token to be used for serverapi connections")

}

func showAoHelp(cmd *cobra.Command, args []string) error {
	cmd.SetHelpTemplate(`{{.Long}}

Usage:
  {{.CommandPath}} [command] [flags]

OpenShift Action Commands:{{range .Commands}}{{if eq (index .Annotations "type") "actions"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Remote AuroraConfig Commands:{{range .Commands}}{{if eq (index .Annotations "type") "remote"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Local File Commands:{{range .Commands}}{{if eq (index .Annotations "type") "local"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Commands:{{range .Commands}}{{if (and (eq (index .Annotations "type") "") (ne .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
	return cmd.Help()
}

func initialize(cmd *cobra.Command, args []string) error {

	// Errors will be printed from main
	cmd.SilenceErrors = true
	// Disable print usage when an error occurs
	cmd.SilenceUsage = true

	home, _ := os.LookupEnv("HOME")
	ConfigLocation = home + "/.ao.json"

	err := setLogging(pFlagLogLevel, pFlagPrettyLog)
	if err != nil {
		return err
	}

	aoConfig, err := config.LoadConfigFile(ConfigLocation)
	if err != nil {
		logrus.Error(err)
	}

	if aoConfig == nil {
		logrus.Info("Creating new config")
		aoConfig = &config.DefaultAOConfig
		aoConfig.InitClusters()
		aoConfig.SelectApiCluster()
		aoConfig.Write(ConfigLocation)
	}

	apiCluster := aoConfig.Clusters[aoConfig.APICluster]
	if apiCluster == nil {
		return errors.Errorf("Api cluster %s is not available. Check config.", aoConfig.APICluster)
	}

	api := client.NewApiClient(apiCluster.BooberUrl, apiCluster.Token, aoConfig.Affiliation)

	if pFlagHost != "" {
		api.Host = pFlagHost
	} else if aoConfig.Localhost {
		// TODO: Move to config?
		api.Host = "http://localhost:8080"
	}

	if pFlagToken != "" {
		api.Token = pFlagToken
	}

	AO, DefaultApiClient = aoConfig, api

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
