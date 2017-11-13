package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/log"
	"github.com/spf13/cobra"
	"os"
)

var (
	logLevel        string
	prettyLog       bool
	persistentHost  string
	persistentToken string
)

// TODO: UPDATE DOCUMENTATION

// DefaultApiClient will use APICluster from ao config as default values
// if persistent token and/or server api url is specified these will override default values
var DefaultApiClient *client.ApiClient

var configLocation string
var ao *config.AOConfig

// TODO: Replace with InitializeOptions
var persistentOptions cmdoptions.CommonCommandOptions

// TODO: Remove all config references
var oldConfig = &configuration.ConfigurationClass{
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
	PersistentPreRunE: Initialize,
}

func init() {
	// TODO: Mark as hidden?
	RootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "", "fatal", "Set loglevel. Valid log levels are [info, debug, warning, error, fatal]")
	RootCmd.PersistentFlags().BoolVarP(&prettyLog, "prettylog", "", false, "Pretty print log")
	RootCmd.PersistentFlags().StringVarP(&persistentHost, "serverapi", "", "", "Override default server API address")
	RootCmd.PersistentFlags().StringVarP(&persistentToken, "token", "", "", "Token to be used for serverapi connections")

	// TODO: Rework
	//setHelpTemplate(RootCmd)
}

func setHelpTemplate(root *cobra.Command) {
	tmp := `{{.Long}}
Usage:
  {{.CommandPath}} [command] [flags]

File Commands:{{range .Commands}}{{if (eq (index .Annotations "type") "file")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Other Commands:{{range .Commands}}{{if (eq (index .Annotations "type") "")}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

{{if .HasPersistentFlags}}
Global Flags:
{{.PersistentFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	root.SetHelpTemplate(tmp)
}

func Initialize(cmd *cobra.Command, args []string) error {

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	home, _ := os.LookupEnv("HOME")
	configLocation = home + "/.ao.json"

	err := setLogging(logLevel, prettyLog)
	if err != nil {
		return err
	}

	ao, err = config.LoadConfigFile(configLocation)
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

	DefaultApiClient = client.NewApiClient(apiCluster.BooberUrl, apiCluster.Token, ao.Affiliation)

	if persistentHost != "" {
		DefaultApiClient.Host = persistentHost
	} else if ao.Localhost {
		// TODO: Move to config?
		DefaultApiClient.Host = "http://localhost:8080"
	}

	if persistentToken != "" {
		DefaultApiClient.Token = persistentToken
	}

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
