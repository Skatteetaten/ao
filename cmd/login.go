// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var userName string
var recreateConfig bool
var apiCluster string
var doUpdate bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Aliases: []string{"affiliation"},
	Use:     "login <Affiliation>",
	Short:   "Login to all available openshift clusters",
	Long: `This command will log in to all available clusters and store the tokens in the .ao.json config file.
If the .ao.json config file does not exist, it will be created.
The command will first check for OpenShift clusters based upon the naming convention implemented by the
NTA.
If these clusters are not found, then the command will use the clusters defined in the OC konfig (cubekonfig).

The --recreate-config flag forces the recreation of .ao.json and will overwrite the previous file.
It is possible to switch API cluster by using the --apicluster flag.

The login command will check for available updates.  The --do-update option will make login do the update if
one is available.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var affiliation string
		if len(args) != 1 {
			fmt.Println("Please specify affiliation to log in to")
			return
		} else {
			affiliation = args[0]
		}

		if recreateConfig {
			conf := &config.DefaultAOConfig
			conf.InitClusters()
			conf.SelectApiCluster()
			ao = conf
		}

		options := config.LoginOptions{
			APICluster:  apiCluster,
			Affiliation: affiliation,
			UserName:    userName,
			LocalHost:   persistentOptions.Localhost,
		}

		ao.Login(configLocation, options)
		// TODO: Check for new ao version
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	user, _ := os.LookupEnv("USER")
	loginCmd.Flags().StringVarP(&userName, "username", "u", user, "the username to log in with, standard is $USER")
	loginCmd.Flags().BoolVarP(&recreateConfig, "recreate-config", "", false, "Removes current cluster config and recreates")
	loginCmd.Flags().StringVarP(&apiCluster, "apicluster", "a", "", "Set a specific API cluster to use")
	loginCmd.Flags().BoolVarP(&doUpdate, "do-update", "", false, "Do an update if available")
}
