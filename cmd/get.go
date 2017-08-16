package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

import pkgGetCmd "github.com/skatteetaten/ao/pkg/getcmd"

var allClusters bool
var getcmdObject pkgGetCmd.GetcmdClass

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves information from the repository",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Usage())
	},
}

var getFileCmd = &cobra.Command{
	Use:     "file",
	Short:   "Get file",
	Aliases: []string{"files"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("file")
	},
}

var getVaultCmd = &cobra.Command{
	Use:     "vault",
	Short:   "Get vault",
	Aliases: []string{"vaults"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("vault")
	},
}

var getSecretCmd = &cobra.Command{
	Use:     "secret <vault> <secret>",
	Short:   "Get secret",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 2 {
			fmt.Println(cmd.UseLine())
			return
		}

		if output, err := getcmdObject.Secret(args[0], args[1], &persistentOptions); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getClusterCmd = &cobra.Command{
	Use:     "cluster",
	Short:   "Get cluster",
	Aliases: []string{"clusters"},
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := ""

		if len(args) > 0 {
			clusterName = args[0]
		}

		if output, err := getcmdObject.Clusters(&persistentOptions, clusterName, allClusters); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getKubeConfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Get kubeconfig",
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := getcmdObject.KubeConfig(&persistentOptions); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getOcLoginCmd = &cobra.Command{
	Use:   "oclogin",
	Short: "Get oclogin",
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := getcmdObject.OcLogin(&persistentOptions); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getFileCmd)
	getCmd.AddCommand(getVaultCmd)
	getCmd.AddCommand(getSecretCmd)
	getCmd.AddCommand(getClusterCmd)
	getCmd.AddCommand(getKubeConfigCmd)
	getCmd.AddCommand(getOcLoginCmd)

	getClusterCmd.Flags().BoolVarP(&allClusters, "all",
		"a", false, "Show all clusters, not just the reachable ones")
}
