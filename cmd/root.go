// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"os"

	"github.com/skatteetaten/aoc/openshift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aoc",
	Short: "Aurora Openshift CLI",
	Long: `A command line interface that interacts with boober

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
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".aoc")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match
	viper.BindEnv("HOME")

	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	openshift.LoadOrInitiateConfig(configLocation)

}
