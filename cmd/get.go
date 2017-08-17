package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	pkgGetCmd "github.com/skatteetaten/ao/pkg/getcmd"
)

var getcmdObject = &pkgGetCmd.GetcmdClass{
	Configuration: config,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves information from the repository",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var getFileCmd = &cobra.Command{
	Use:     "file [envname] <filename>",
	Short:   "Get file",
	Aliases: []string{"files"},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		if len(args) == 0 {
			output, err = getcmdObject.Files()
		} else {
			output, err = getcmdObject.File(args)
		}

		if err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getVaultCmd = &cobra.Command{
	Use:     "vault [vaultname]",
	Short:   "Get vault",
	Aliases: []string{"vaults"},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		if len(args) == 0 {
			output, err = getcmdObject.Vaults()
		} else {
			output, err = getcmdObject.Vault(args[0])
		}

		if err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getSecretCmd = &cobra.Command{
	Use:   "secret <vault> <secret>",
	Short: "Get secret",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 2 {
			fmt.Println(cmd.UseLine())
			return
		}

		if output, err := getcmdObject.Secret(args[0], args[1]); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getClusterCmd = &cobra.Command{
	Use:     "cluster [clustername]",
	Short:   "Get cluster",
	Aliases: []string{"clusters"},
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := ""

		if len(args) > 0 {
			clusterName = args[0]
		}

		allClusters, _ := cmd.Flags().GetBool("all")
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

	getClusterCmd.Flags().BoolP("all",
		"a", false, "Show all clusters, not just the reachable ones")
}
