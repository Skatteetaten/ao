package cmd

import (
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/spf13/cobra"
)

var deployOptions command.DeployOptions

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

		envList, _ := cmd.Flags().GetStringArray("env")
		appList, _ := cmd.Flags().GetStringArray("app")

		allArgs := append(envList, appList...)
		allArgs = append(allArgs, args...)

		if len(allArgs) < 1 && !deployOptions.DeployAll {
			cmd.Usage()
			return
		}

		options := &deployOptions
		if options.Affiliation == "" {
			options.Affiliation = ao.Affiliation
		}

		options.DeployOnce = ao.Localhost || options.DeployApiOnly || options.Cluster != ""
		options.Token = persistentOptions.Token

		result := command.Deploy(allArgs, DefaultApiClient, ao.Clusters, options)
		if result != nil {
			command.PrintDeployResults(result)
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringArrayP("app",
		"a", nil, "Only deploy specified application")

	deployCmd.Flags().StringArrayP("env",
		"e", nil, "Only deploy specified environment")

	deployCmd.Flags().StringArrayVarP(&deployOptions.Overrides, "overrides",
		"o", []string{}, "Override in the form [env/]file:{<json override>}")

	deployCmd.Flags().BoolVarP(&deployOptions.DeployAll, "all",
		"", false, "Will deploy all applications in all clusters reachable")

	deployCmd.Flags().BoolVarP(&deployOptions.Force, "force",
		"", false, "Supress prompts")

	deployCmd.Flags().StringVarP(&deployOptions.Version, "version",
		"v", "", "Will update the version tag in the app of base configuration file prior to deploy, depending on which file contains the version tag.  If both files "+
			"files contains the tag, the tag will be updated in the app configuration file.")

	deployCmd.Flags().StringVarP(&deployOptions.Affiliation, "affiliation",
		"", "", "Overrides the logged in affiliation")

	deployCmd.Flags().StringVarP(&deployOptions.Cluster, "cluster", "c", "",
		"Limit deploy to given clustername")

	deployCmd.Flags().BoolVarP(&deployOptions.DeployApiOnly, "api-cluster-only", "", false,
		"Limit deploy to the API cluster")
}
