package cmd

import (
	"fmt"
	config2 "github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Perform administrative commands on AO or other resources not related to vaults of AuroraConfig",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var getClusterCmd = &cobra.Command{
	Use:   "cluster [clustername]",
	Short: "Get cluster",
	Long: `The command will list the reachable OpenShift clusters defined in the configuration file (~/ao.json).
If the --all flag is specified, all clusters will be listed.
The API cluster is the one used to serve configuration requests.  All the commands except Deploy will only use the
API cluster.
The Deploy command will send the same request to all the reachable clusters, allowing each to filter deploys
intended for that particular cluster.`,
	Aliases: []string{"clusters"},
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := ""

		if len(args) > 0 {
			clusterName = args[0]
		}

		allClusters, _ := cmd.Flags().GetBool("all")
		ao.PrintClusters(clusterName, allClusters)
	},
}

var updateClustersCmd = &cobra.Command{
	Use:   "update-clusters",
	Short: "Will update clusters",
	Long:  `The command will updated cluster with latest clusterUrlPattern and booberUrlPattern.`,
	Run: func(cmd *cobra.Command, args []string) {
		ao.InitClusters()
		ao.SelectApiCluster()
		ao.Write(configLocation)
	},
}

var recreateConfigCmd = &cobra.Command{
	Use:   "recreate-config",
	Short: "adm recreate-config",
	Long:  `The command will recreate the .ao.json file.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := &config2.DefaultAOConfig
		conf.InitClusters()
		conf.SelectApiCluster()

		cluster, _ := cmd.Flags().GetString("cluster")
		if cluster != "" {
			conf.AvailableClusters = []string{cluster}
		}
		conf.Write(configLocation)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion file",
	Long: `This command generates a bash script file that provides bash completion.
Bash completion allows you to press the Tab key to complete keywords.
After running this command, a file called ao.sh will exist in your home directory.
By typing
	source ./ao.sh
you will have bash completion in ao
To persist this across login sessions, please update your .bashrc file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RootCmd.GenBashCompletionFile("ao.sh"); err == nil {
			wd, _ := os.Getwd()
			fmt.Println("Bash completion file created at", wd+"/ao.sh")
		} else {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(admCmd)
	admCmd.AddCommand(getClusterCmd)
	admCmd.AddCommand(completionCmd)
	admCmd.AddCommand(recreateConfigCmd)
	admCmd.AddCommand(updateClustersCmd)

	// Get cluster
	getClusterCmd.Flags().BoolP("all", "a", false, "Show all clusters, not just the reachable ones")

	// Recreate config
	recreateConfigCmd.Flags().StringP("cluster", "c", "", "Limit recreate-config to the given Tax Norway cluster")
}
