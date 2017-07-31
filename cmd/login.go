// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/skatteetaten/aoc/pkg/updatecmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var userName string
var tokenFile string
var recreateConfig bool
var useCurrentOcLogin bool
var apiCluster string
var doUpdate bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <Affiliation>",
	Short: "Login to openshift clusters",
	Long:  `This command will log in to all avilable clusters and store the tokens in the .aoc config file `,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please specify affiliation to log in to")
			os.Exit(1)
		}
		affiliation := args[0]
		var configLocation = viper.GetString("HOME") + "/.aoc.json"
		if recreateConfig || useCurrentOcLogin {
			err := os.Remove(configLocation)
			if err != nil {
				if !strings.Contains(err.Error(), "no such file or directory") {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}
		}
		initConfig(useCurrentOcLogin)
		openshift.Login(configLocation, userName, affiliation, apiCluster)
		output, _ := updatecmd.UpdateSelf(args, !doUpdate, "", false)
		if strings.Contains(output, "New version detected") {
			fmt.Println(output)
		}
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
	viper.BindEnv("USER")
	viper.BindEnv("HOME")
	loginCmd.Flags().StringVarP(&userName, "username", "u", viper.GetString("USER"), "the username to log in with, standard is $USER")
	loginCmd.Flags().StringVarP(&tokenFile, "tokenfile", "", "", "Read OC token from this file")
	loginCmd.Flags().BoolVarP(&recreateConfig, "recreate-config", "", false, "Removes current cluster config and recreates")
	loginCmd.Flags().BoolVarP(&useCurrentOcLogin, "use-current-oclogin", "", false, "Recreates config based on current OC login")
	loginCmd.Flags().StringVarP(&apiCluster, "apicluster", "a", "", "Set a specific API cluster to use")
	loginCmd.Flags().BoolVarP(&doUpdate, "do-update", "", false, "Do an update if available")
}
