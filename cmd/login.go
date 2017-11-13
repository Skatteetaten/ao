// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

const loginLong = `This command will log in to all available clusters and store the tokens in the .ao.json config file.
If the .ao.json config file does not exist, it will be created.
The command will first check for OpenShift clusters based upon the naming convention implemented by the
NTA.
If these clusters are not found, then the command will use the clusters defined in the OC konfig (cubekonfig).

The --recreate-config flag forces the recreation of .ao.json and will overwrite the previous file.
It is possible to switch API cluster by using the --apicluster flag.

The login command will check for available updates.  The --do-update option will make login do the update if
one is available.
`

var (
	userName       string
	recreateConfig bool
	apiCluster     string
	doUpdate       bool
	localhost      bool
)

var loginCmd = &cobra.Command{
	Aliases: []string{"affiliation"},
	Use:     "login <Affiliation>",
	Short:   "Login to all available openshift clusters",
	Long:    loginLong,
	RunE:    Login,
}

func init() {
	RootCmd.AddCommand(loginCmd)
	user, _ := os.LookupEnv("USER")
	loginCmd.Flags().StringVarP(&userName, "username", "u", user, "the username to log in with, standard is $USER")
	loginCmd.Flags().BoolVarP(&recreateConfig, "recreate-config", "", false, "Removes current cluster config and recreates")
	loginCmd.Flags().StringVarP(&apiCluster, "apicluster", "a", "", "Set a specific API cluster to use")
	loginCmd.Flags().BoolVarP(&doUpdate, "do-update", "", false, "Do an update if available")
	loginCmd.Flags().BoolVarP(&localhost, "localhost", "l", false, "Development mode")
	loginCmd.Flags().MarkHidden("localhost")
}

func Login(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Please specify affiliation to log in to")
	}

	if recreateConfig {
		conf := &config.DefaultAOConfig
		conf.InitClusters()
		conf.SelectApiCluster()
		ao = conf
	}

	options := config.LoginOptions{
		APICluster:  apiCluster,
		Affiliation: args[0],
		UserName:    userName,
		LocalHost:   localhost,
	}

	ao.Login(configLocation, options)
	err := ao.Update()
	if err != nil {
		return err
	}

	fmt.Println("AO has been updated")
	return nil
}
