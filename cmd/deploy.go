// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
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
	"log"
	"os"

	"github.com/skatteetaten/ao/pkg/deploy"
	"github.com/spf13/cobra"
	"github.com/stromland/coprompt"
)

var appList []string
var envList []string
var overrideJson []string
var deployAllFlag bool
var forceDeployFlag bool
var deployVersion string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy applications in the current affiliation",
	Long: `Deploy applications in the current affiliation.

A Deploy will compare the stored configuration with the running projects in OpenShift, and update the OpenShift
environment to match the specifications in the stored configuration.

If no changes is detected, no updates to OpenShift will be done (except for an update of the resourceVersion in the BuildConfig).

As per default, all applications in all environments will be deployed.
Using the -e flag, it is possible to limit the deploy to the specified environment.
Using the -a flag, it is possible to limit the deploy to the specified application.
Both flags can be used to limit the deploy to a specific application in a specific environment.`,
	Aliases: []string{"setup"},
	Annotations: map[string]string{
		coprompt.CALLBACK_ANNOTATION: "GetDeployments",
	},
	Run: func(cmd *cobra.Command, args []string) {
		deploy := deploy.DeployClass{
			Configuration: config,
		}

		output, err := deploy.ExecuteDeploy(args, overrideJson, appList, envList, &persistentOptions, localDryRun, deployAllFlag, forceDeployFlag, deployVersion)
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
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringArrayVarP(&overrideJson, "file",
		"o", overrideValues, "Override in the form [env/]file:{<json override>}")

	deployCmd.Flags().BoolVarP(&localDryRun, "localdryrun",
		"z", false, "Does not initiate API, just prints collected files")
	deployCmd.Flags().MarkHidden("localdryrun")

	deployCmd.Flags().StringArrayVarP(&appList, "app",
		"a", nil, "Only deploy specified application")

	deployCmd.Flags().StringArrayVarP(&envList, "env",
		"e", nil, "Only deploy specified environment")

	deployCmd.Flags().BoolVarP(&deployAllFlag, "all",
		"", false, "Will deploy all applications in all affiliations in all clusters reachable")

	deployCmd.Flags().BoolVarP(&forceDeployFlag, "force",
		"", false, "Supress prompts")

	deployCmd.Flags().StringVarP(&deployVersion, "version",
		"v", "", "Will update the version tag in the app of base configuration file prior to deploy, depending on which file contains the version tag.  If both files "+
			"files contains the tag, the tag will be updated in the app configuration file.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// File flag, supports multiple instances of the flag

}
