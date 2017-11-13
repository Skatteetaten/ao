package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Perform administrative commands on AO or other resources not related to vaults of AuroraConfig",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
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
	Run:     PrintClusters,
}

var updateClustersCmd = &cobra.Command{
	Use:   "update-clusters",
	Short: "Will update clusters",
	Long:  `The command will updated cluster with latest clusterUrlPattern and booberUrlPattern.`,
	Run:   UpdateClusters,
}

var recreateConfigCmd = &cobra.Command{
	Use:   "recreate-config",
	Short: "adm recreate-config",
	Long:  `The command will recreate the .ao.json file.`,
	Run:   RecreateConfig,
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
	RunE: BashCompletion,
}

func init() {
	RootCmd.AddCommand(admCmd)
	admCmd.AddCommand(getClusterCmd)
	admCmd.AddCommand(completionCmd)
	admCmd.AddCommand(recreateConfigCmd)
	admCmd.AddCommand(updateClustersCmd)

	getClusterCmd.Flags().BoolP("all", "a", false, "Show all clusters, not just the reachable ones")
}

func PrintClusters(cmd *cobra.Command, args []string) {
	allClusters, _ := cmd.Flags().GetBool("all")
	table := []string{"\tCLUSTER NAME\tREACHABLE\tLOGGED IN\tAPI\tURL"}

	for name, cluster := range ao.Clusters {
		if !(cluster.Reachable || allClusters) {
			continue
		}
		reachable := ""
		if cluster.Reachable {
			reachable = "Yes"
		}

		loggedIn := ""
		if cluster.HasValidToken() {
			loggedIn = "Yes"
		}

		api := ""
		if name == ao.APICluster {
			api = "Yes"
		}
		line := fmt.Sprintf("\t%s\t%s\t%s\t%s\t%s\t", name, reachable, loggedIn, api, cluster.Url)
		table = append(table, line)
	}

	DefaultTablePrinter(table)
}

func UpdateClusters(cmd *cobra.Command, args []string) {
	ao.InitClusters()
	ao.SelectApiCluster()
	ao.Write(configLocation)
}

func RecreateConfig(cmd *cobra.Command, args []string) {
	conf := &config.DefaultAOConfig
	conf.InitClusters()
	conf.SelectApiCluster()

	cluster, _ := cmd.Flags().GetString("cluster")
	if cluster != "" {
		conf.AvailableClusters = []string{cluster}
	}
	conf.Write(configLocation)
}

func BashCompletion(cmd *cobra.Command, args []string) error {
	err := RootCmd.GenBashCompletionFile("ao.sh")
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	fmt.Println("Bash completion file created at", wd+"/ao.sh")
	return nil
}
