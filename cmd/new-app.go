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
var newappFolder string
var newappGenerateApp bool

// newAppCmd represents the newApp command
var newAppCmd = &cobra.Command{
	Use:   "new-app <appname>",
	Short: "Creates an AuroraConfig for an application and optionally generates an an empty application",
	Long: `Creates an AuroraConfig for an application and uploads it to Boober.
As default, the generate-app flag is set to true.  This will call the Yeoman generator for creating applications to run on the Aurora OpenShift platform.
The Yeoman generator will create the application files in the current folder.  The folder to be used can be overridden with the --folder / -f flag.
If the folder is not empty, the command will return an error.

If the generate-app is set to false, the generator will not be called, and the command will generate a set of Aurora config files based upon the arguments.

If the artifactid is not given, it will default to the appname.`,
	Run: func(cmd *cobra.Command, args []string) {
		var newappcmdObject newappcmd.NewappcmdClass
		output, err := newappcmdObject.NewappCommand(args, newappArtifactId, newappCluster, newappEnv, newappGroupId, newappFolder, newappOutputfolder, newappType, newappVersion, newappGenerateApp, &persistentOptions)
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

	newAppCmd.Flags().BoolVarP(&newappGenerateApp, "generate-app", "", true, "Calls the Yeoman generator to generate a sample app")
	newAppCmd.Flags().StringVarP(&newappFolder, "folder", "f", ".", "Specifies the folder name for the Yeoman generator.  Calls the generator to prompt for input, and then reads the values from the config files.")
	newAppCmd.Flags().StringVarP(&newappType, "type", "t", "development", "Type of deploy: development or deploy")
	newAppCmd.Flags().StringVarP(&newappGroupId, "groupid", "g", "", "GroupID for the application")
	newAppCmd.Flags().StringVarP(&newappArtifactId, "artifactid", "a", "", "Artifact ID for the application")
	newAppCmd.Flags().StringVarP(&newappVersion, "version", "v", "latest", "Version for the application")
	newAppCmd.Flags().StringVarP(&newappCluster, "cluster", "c", "", "OpenShift Clustername target, defaults to AOC API cluster")
	newAppCmd.Flags().StringVarP(&newappEnv, "env", "e", viper.GetString("USER"), "Environment folder for the config, defaults to username")
	newAppCmd.Flags().StringVarP(&newappOutputfolder, "output-folder", "o", "", "If specified the files are generated in this folder instead of being sent to Boober")
}
