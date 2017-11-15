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
	flagUserName       string
	flagRecreateConfig bool
	flagApiCluster     string
	flagNoUpdatePrompt bool
	flagLocalhost      bool
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
	loginCmd.Flags().StringVarP(&flagUserName, "username", "u", user, "the username to log in with, standard is $USER")
	loginCmd.Flags().BoolVarP(&flagRecreateConfig, "recreate-config", "", false, "Removes current cluster config and recreates")
	loginCmd.Flags().StringVarP(&flagApiCluster, "apicluster", "a", "", "Set a specific API cluster to use")
	loginCmd.Flags().BoolVarP(&flagNoUpdatePrompt, "do-update", "", false, "Do an update if available")
	loginCmd.Flags().BoolVarP(&flagLocalhost, "localhost", "l", false, "Development mode")
	loginCmd.Flags().MarkHidden("localhost")
}

func Login(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Please specify affiliation to log in to")
	}

	if flagRecreateConfig {
		conf := &config.DefaultAOConfig
		conf.InitClusters()
		conf.SelectApiCluster()
		AO = conf
	}

	options := config.LoginOptions{
		APICluster:  flagApiCluster,
		Affiliation: args[0],
		UserName:    flagUserName,
		LocalHost:   flagLocalhost,
	}

	AO.Login(ConfigLocation, options)
	err := AO.Update(flagNoUpdatePrompt)
	if err != nil {
		return err
	}

	fmt.Println("AO has been updated")
	return nil
}
