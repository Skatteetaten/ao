// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stromland/cobra-prompt"
)

const CallbackAnnotation = cobraprompt.CALLBACK_ANNOTATION

var persistentOptions cmdoptions.CommonCommandOptions
var overrideValues []string
var localDryRun bool

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
		commandsWithoutLogin := []string{"login", "logout", "version", "help"}

		commands := strings.Split(cmd.CommandPath(), " ")
		if len(commands) > 1 {
			for _, command := range commandsWithoutLogin {
				if commands[1] == command {
					return
				}
			}
		}

		config.OpenshiftConfig = getOpenShiftConfig(false, "")

		if err := config.SetApiCluster(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if valid := serverapi.ValidateLogin(config.OpenshiftConfig); !valid {
			fmt.Println("Not logged in, please use ao login")
			os.Exit(1)
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

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Verbose, "verbose",
		"", false, "Log progress to standard out")

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Debug, "debug",
		"", false, "Show debug information")
	RootCmd.PersistentFlags().MarkHidden("debug")
	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Localhost, "localhost",
		"l", false, "Send setup to Boober on localhost")
	RootCmd.PersistentFlags().MarkHidden("localhost")

	RootCmd.PersistentFlags().StringVarP(&persistentOptions.ServerApi, "serverapi",
		"", "", "Override default server API address")
	RootCmd.PersistentFlags().StringVarP(&persistentOptions.Token, "token",
		"", "", "Token to be used for serverapi connections")
}

func getOpenShiftConfig(useOcConfig bool, loginCluster string) *openshift.OpenshiftConfig {
	viper.SetConfigName(".ao")   // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match
	viper.BindEnv("HOME")

	aoConfigLocation := viper.GetString("HOME") + "/.ao.json"
	openShiftConfig, _ := openshift.LoadOrInitiateConfigFile(aoConfigLocation, loginCluster, useOcConfig)
	openShiftConfig.ConfigLocation = aoConfigLocation

	return openShiftConfig
}
