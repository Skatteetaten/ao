package cmd

import (
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/spf13/cobra"
)

var appList []string
var envList []string
var overrideJson []string
var deployAllFlag bool
var forceDeployFlag bool

//var deployVersion string
var deployAffiliation string
var deployCluster string
var deployApiClusterOnly bool

var deployLong = `Deploy applications for the current affiliation.

A Deploy will compare the stored configuration with the running projects in OpenShift, and update the OpenShift
environment to match the specifications in the stored configuration.

If no changes is detected, no updates to OpenShift will be done (except for an update of the resourceVersion in the BuildConfig).

In addition, the command accepts a mixed list of applications and environments on the command line.
The names may be shortened; the command will search the current affiliation for unique matches.

If the command will result in multiple deploys, a confirmation dialog will be shown, listing the result of the command.
The list will contain all the affected applications and environments.  Please note that the two columns are not correlated.
The --force flag will override this, and execute the deploy without confirmation.

`
var deployCmd = &cobra.Command{
	Aliases: []string{"setup"},
	Use:     "deploy",
	Short:   "Deploy applications for the current affiliation",
	Long:    deployLong,
	Run: func(cmd *cobra.Command, args []string) {

		allArgs := append(envList, appList...)
		allArgs = append(allArgs, args...)

		if len(allArgs) < 1 && !deployAllFlag {
			cmd.Usage()
			return
		}

		affiliation := ao.Affiliation
		if deployAffiliation != "" {
			affiliation = deployAffiliation
		}

		deployOnce := ao.Localhost || deployApiClusterOnly || deployCluster != ""

		options := command.DeployOptions{
			Affiliation: affiliation,
			Overrides:   overrideJson,
			Force:       forceDeployFlag,
			DeployAll:   deployAllFlag,
			DeployOnce:  deployOnce,
			Cluster:     deployCluster,
			Token:       persistentOptions.Token,
		}

		command.Deploy(allArgs, DefaultApiClient, ao.Clusters, options)
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringArrayVarP(&overrideJson, "overrides",
		"o", []string{}, "Override in the form [env/]file:{<json override>}")

	deployCmd.Flags().StringArrayVarP(&appList, "app",
		"a", nil, "Only deploy specified application")

	deployCmd.Flags().StringArrayVarP(&envList, "env",
		"e", nil, "Only deploy specified environment")

	deployCmd.Flags().BoolVarP(&deployAllFlag, "all",
		"", false, "Will deploy all applications in all clusters reachable")

	deployCmd.Flags().BoolVarP(&forceDeployFlag, "force",
		"", false, "Supress prompts")

	// TODO
	//deployCmd.Flags().StringVarP(&deployVersion, "version",
	//	"v", "", "Will update the version tag in the app of base configuration file prior to deploy, depending on which file contains the version tag.  If both files "+
	//		"files contains the tag, the tag will be updated in the app configuration file.")

	deployCmd.Flags().StringVarP(&deployAffiliation, "affiliation",
		"", "", "Overrides the logged in affiliation")

	deployCmd.Flags().StringVarP(&deployCluster, "cluster", "c", "",
		"Limit deploy to given clustername")

	deployCmd.Flags().BoolVarP(&deployApiClusterOnly, "api-cluster-only", "", false,
		"Limit deploy to the API cluster")
}
