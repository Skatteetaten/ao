// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"github.com/skatteetaten/aoc/openshift"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	urlPattern = "https://%s-master.paas.skead.no:8443"
)

var userName string

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to openshift clusters",
	Long:  `This command will log in to all avilable clusters and store the tokens in the .aoc config file `,
	Run: func(cmd *cobra.Command, args []string) {
		var configLocation = viper.GetString("HOME") + "/.aoc.json"
		openshift.LoginToAllCluster(configLocation, userName)
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	viper.BindEnv("USER")
	viper.BindEnv("HOME")
	loginCmd.LocalFlags().StringVarP(&userName, "username", "u", viper.GetString("USER"), "the username to log in with, standard is $USER")
}
