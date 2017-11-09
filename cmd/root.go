package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/spf13/cobra"
	"os"
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		home, _ := os.LookupEnv("HOME")
		configLocation = home + "/.ao.json"

		ao, DefaultApiClient = command.Initialize(configLocation, command.InitializeOptions{
			Host:        persistentOptions.ServerApi,
			Token:       persistentOptions.Token,
			LogLevel:    persistentOptions.LogLevel,
			PrettyLog:   persistentOptions.Pretty,
			CommandName: cmd.Name(),
			CommandPath: cmd.CommandPath(),
		})
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
