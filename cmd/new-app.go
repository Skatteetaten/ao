// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/skatteetaten/aoc/pkg/newappcmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var newappType string
var newappGroupId string
var newappArtifactId string = ""
var newappVersion string
var newappCluster string
var newappName string
var newappEnv string
var newappOutputfolder string
var newappInteractive string

// newAppCmd represents the newApp command
var newAppCmd = &cobra.Command{
	Use:   "new-app <appname>",
	Short: "Creates an AuroraConfig for an application",
	Long: `Creates an AuroraConfig for an application and uploads it to Boober.
The appname parameter is optional if artifactid is set; if not specified the app will be named as the artifactid.
If the artifactid is not given, it will default to the appname.`,
	Run: func(cmd *cobra.Command, args []string) {
		var newappcmdObject newappcmd.NewappcmdClass
		output, err := newappcmdObject.NewappCommand(args, newappArtifactId, newappCluster, newappEnv, newappGroupId, newappInteractive, newappOutputfolder, newappType, newappVersion, &persistentOptions)
		if err != nil {
			l := log.New(os.Stderr, "", 0)
			l.Println(err.Error())
			os.Exit(-1)
		} else {
			if output != "" {
				fmt.Println(output)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(newAppCmd)
	viper.BindEnv("USER")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newAppCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newAppCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	newAppCmd.Flags().StringVarP(&newappInteractive, "interactive", "i", "", "Specifies the folder name for the Yeoman generator.  Calls the generator to prompt for input, and then reads the values from the config files.")
	newAppCmd.Flags().StringVarP(&newappType, "type", "t", "development", "Type of deploy: development or deploy")
	newAppCmd.Flags().StringVarP(&newappGroupId, "groupid", "g", "", "GroupID for the application")
	newAppCmd.Flags().StringVarP(&newappArtifactId, "artifactid", "a", "", "Artifact ID for the application")
	newAppCmd.Flags().StringVarP(&newappVersion, "version", "v", "latest", "Version for the application")
	newAppCmd.Flags().StringVarP(&newappCluster, "cluster", "c", "", "OpenShift Clustername target, defaults to AOC API cluster")
	newAppCmd.Flags().StringVarP(&newappEnv, "env", "e", viper.GetString("USER"), "Environment folder for the config, defaults to username")
	newAppCmd.Flags().StringVarP(&newappOutputfolder, "output-folder", "o", "", "If specified the files are generated in this folder instead of being sent to Boober")
}
