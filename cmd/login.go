// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/updatecmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var userName string
var recreateConfig bool
var useCurrentOcLogin bool
var apiCluster string
var doUpdate bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <Affiliation>",
	Short: "Login to all available openshift clusters",
	Long: `This command will log in to all available clusters and store the tokens in the .ao.json config file.
If the .ao.json config file does not exist, it will be created.
The command will first check for OpenShift clusters based upon the naming convention implemented by the
NTA.
If these clusters are not found, then the command will use the clusters defined in the OC konfig (cubekonfig).

The --recreate-config flag forces the recreation of .ao.json and will overwrite the previous file.
The --use-current-oclogin will force the creation of config based upon the OC config, even in a NTA environment.
It is possible to switch API cluster by using the --apicluster flag.

The login command will check for available updates.  The --do-update option will make login do the update if
one is available.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var affiliation string
		if len(args) != 1 {
			if !recreateConfig && !useCurrentOcLogin { // && !recreateConfig && !useCurrentOcLogin
				fmt.Println("Please specify affiliation to log in to")
				os.Exit(1)
			}
		} else {
			affiliation = args[0]
		}
		var configLocation = viper.GetString("HOME") + "/.ao.json"
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
		if !recreateConfig && !useCurrentOcLogin {
			openshift.Login(configLocation, userName, affiliation, apiCluster, persistentOptions.Localhost)
		}
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
	//loginCmd.Flags().StringVarP(&tokenFile, "tokenfile", "", "", "Read OC token from this file")
	loginCmd.Flags().BoolVarP(&recreateConfig, "recreate-config", "", false, "Removes current cluster config and recreates")
	loginCmd.Flags().BoolVarP(&useCurrentOcLogin, "use-current-oclogin", "", false, "Recreates config based on current OC login")
	loginCmd.Flags().StringVarP(&apiCluster, "apicluster", "a", "", "Set a specific API cluster to use")
	loginCmd.Flags().BoolVarP(&doUpdate, "do-update", "", false, "Do an update if available")
}
