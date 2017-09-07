package cmd

import (
	"fmt"
	"os"
	"strings"

	pkgGetCmd "github.com/skatteetaten/ao/pkg/getcmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var admcmdObject = &pkgGetCmd.GetcmdClass{
	Configuration: config,
}

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
		if output, err := admcmdObject.Clusters(clusterName, allClusters); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getKubeConfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "adm kubeconfig",
	Long:  `The command will list the contents of the OC configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := admcmdObject.KubeConfig(); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getOcLoginCmd = &cobra.Command{
	Use:   "oclogin",
	Short: "adm oclogin",
	Long:  `The command will print info about the current OC login.`,
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := admcmdObject.OcLogin(); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var recreateConfigCmd = &cobra.Command{
	Use:   "recreate-config",
	Short: "adm recreate-config",
	Long:  `The command will recreate the .ao.json file.`,
	Run: func(cmd *cobra.Command, args []string) {

		var configLocation = viper.GetString("HOME") + "/.ao.json"
		err := os.Remove(configLocation)
		if err != nil {
			if !strings.Contains(err.Error(), "no such file or directory") {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		initConfig(useCurrentOcLogin)
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
	admCmd.AddCommand(getKubeConfigCmd)
	admCmd.AddCommand(getOcLoginCmd)
	admCmd.AddCommand(completionCmd)

	getClusterCmd.Flags().BoolP("all",
		"a", false, "Show all clusters, not just the reachable ones")

	admCmd.AddCommand(recreateConfigCmd)
	recreateConfigCmd.Flags().BoolVarP(&useCurrentOcLogin, "use-current-oclogin", "", false, "Recreates config based on current OC login")
}
