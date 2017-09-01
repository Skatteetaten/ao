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
)

var appList []string
var envList []string
var overrideJson []string
var deployAllFlag bool
var forceDeployFlag bool
var deployVersion string
var deployAffiliation string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy applications in the current affiliation",
	Long: `Deploy applications in the current affiliation.

A Deploy will compare the stored configuration with the running projects in OpenShift, and update the OpenShift
environment to match the specifications in the stored configuration.

If no changes is detected, no updates to OpenShift will be done (except for an update of the resourceVersion in the BuildConfig).

Using the -e flag, it is possible to limit the deploy to the specified environment.
Using the -a flag, it is possible to limit the deploy to the specified application.
Both flags can be used to limit the deploy to a specific application in a specific environment.

The --all flag will deploy all applications in all environements.

In addition, the command accepts a mixed list of applications and environments on the command line.
The names may be shortened; the command will search the current affiliation for unique matches.

If you have 2 environments named superapp-test and superapp-prod, both containing the applications superapp and niceapp,
then the command

	ao deploy test

will deploy superapp and niceapp in the superapp-test environment.

The command

	ao deploy nice pro

will deploy niceapp in the superapp-prod environment.

However, the command

	ao deploy superapp

will fail, because superapp match both an application and an environment.  Use the -a og -e flag to specify.

It is also possible to specify the env and app in the form app/env, so the command

	ao deploy app-test/nic

will deploy niceapp in the superapp-test environment.

If the command will result in multiple deploys, a confirmation dialog will be shown, listing the result of the command.
The list will contain all the affected applications and environments.  Please note that the two columns are not correlated.
The --force flag will override this, and execute the deploy without confirmation.

`,
	Aliases: []string{"setup"},
	Annotations: map[string]string{
		CallbackAnnotation: "GetDeployments",
	},
	Run: func(cmd *cobra.Command, args []string) {
		deploy := deploy.DeployClass{
			Configuration: config,
		}

		output, err := deploy.ExecuteDeploy(args, overrideJson, appList, envList, &persistentOptions, localDryRun, deployAllFlag, forceDeployFlag, deployVersion, deployAffiliation)
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

	deployCmd.Flags().StringVarP(&deployAffiliation, "affiliation",
		"", "", "Overrides the logged in affiliation")
}
