// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// Cobra Flag variables
var persistentOptions cmdoptions.CommonCommandOptions
var overrideFiles []string
var overrideValues []string
var localDryRun bool

//var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ao",
	Short: "Aurora Openshift CLI",
	Long: `A command line interface that interacts with the Boober serverapi
to enable the user to manipulate the Aurora Config for an affiliation, and to
 deploy one or more application.

This application has two main parts.
1. manage the aoc configuration via cli
2. apply the aoc configuration to the clusters
`,
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
	cobra.OnInitialize(initConfigCobra)

	//Verbose     bool
	//Debug       bool
	//DryRun      bool
	//Localhost   bool
	//ShowConfig  bool
	//ShowObjects bool

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Verbose, "verbose",
		"", false, "Log progress to standard out")

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Debug, "debug",
		"", false, "Show debug information")
	RootCmd.PersistentFlags().MarkHidden("debug")

	//RootCmd.PersistentFlags().BoolVarP(&persistentOptions.DryRun, "dryrun",
	//	"d", false,
	//	"Do not perform a setup, just collect and print the configuration files")

	RootCmd.PersistentFlags().BoolVarP(&persistentOptions.Localhost, "localhost",
		"l", false, "Send setup to Boober on localhost")
	RootCmd.PersistentFlags().MarkHidden("localhost")

	RootCmd.PersistentFlags().StringVarP(&persistentOptions.ServerApi, "serverapi",
		"", "", "Override default server API address")
	//RootCmd.PersistentFlags().MarkHidden("serverurl")
	RootCmd.PersistentFlags().StringVarP(&persistentOptions.Token, "token",
		"", "", "Token to be used for serverapi connections")

	//RootCmd.PersistentFlags().BoolVarP(&persistentOptions.ShowConfig, "showconfig",
	//	"", false, "Print merged config from Boober to standard out")

	//RootCmd.PersistentFlags().BoolVarP(&persistentOptions.ShowObjects, "showobjects",
	//	"", false, "Print object definitions from Boober to standard out")
	// test

}

// initConfig reads in config file and ENV variables if set.

func initConfigCobra() {
	initConfig(false)
}

func initConfig(useOcConfig bool) {
	viper.SetConfigName(".ao")   // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match
	viper.BindEnv("HOME")

	var configLocation = viper.GetString("HOME") + "/.ao.json"
	openshift.LoadOrInitiateConfigFile(configLocation, useOcConfig)

}
