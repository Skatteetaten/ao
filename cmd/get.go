package cmd

import (
	"fmt"

	pkgGetCmd "github.com/skatteetaten/ao/pkg/getcmd"
	"github.com/spf13/cobra"
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

var getDeploymentsCmd = &cobra.Command{
	Use:     "deployment",
	Short:   "get deployments",
	Long:    `Lists the deployments defined in the Auroraconfig`,
	Aliases: []string{"deployments", "dep", "deps", "all"},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		output, err = getcmdObject.Deployments("")

		if err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getAppsCmd = &cobra.Command{
	Use:     "app",
	Short:   "get app",
	Long:    `Lists the apps defined in the Auroraconfig`,
	Aliases: []string{"apps"},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		if len(args) == 0 {
			output, err = getcmdObject.Apps()
		} else {
			output, err = getcmdObject.Deployments(args[0])
		}
		if err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getEnvsCmd = &cobra.Command{
	Use:     "env",
	Short:   "get env",
	Long:    `Lists the envs defined in the Auroraconfig`,
	Aliases: []string{"envs"},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		output, err = getcmdObject.Envs()

		if err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getFileCmd = &cobra.Command{
	Use:   "file [envname] <filename>",
	Short: "Get file",
	Long: `Prints the content of the file to standard output.
Environmentnames and filenames can be abbrevated, and can be specified either as separate strings,
or on a env/file basis.

Given that a file called superapp-test/about.json exists in the repository, the command

	ao get file test ab

will print the file.

If no argument is given, the command will list all the files in the repository.`,
	Aliases: []string{"files"},
	Annotations: map[string]string{
		CallbackAnnotation: "GetFiles",
	},
	Run: func(cmd *cobra.Command, args []string) {

		var output string
		var err error

		if len(args) == 0 {
			output, err = getcmdObject.Files()
		} else {
			output, err = getcmdObject.File(args)
		}

		if err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getVaultCmd = &cobra.Command{
	Use:   "vault [vaultname]",
	Short: "Get vault",
	Long: `If no argument is given, the command will list the vaults in the current affiliation, along with the
numer of secrets in the vault.
If a vaultname is specified, the command will list the secrets in the given vault.
To access a secret, use the get secret command.`,
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
	Long:  `The command will print the content of the secret to standard out.`,
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
		if output, err := getcmdObject.Clusters(clusterName, allClusters); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getKubeConfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Get kubeconfig",
	Long:  `The command will list the contents of the OC configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := getcmdObject.KubeConfig(); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var getOcLoginCmd = &cobra.Command{
	Use:   "oclogin",
	Short: "Get oclogin",
	Long:  `The command will print info about the current OC login.`,
	Run: func(cmd *cobra.Command, args []string) {
		if output, err := getcmdObject.OcLogin(); err == nil {
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
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getEnvsCmd)
	getCmd.AddCommand(getDeploymentsCmd)

	getClusterCmd.Flags().BoolP("all",
		"a", false, "Show all clusters, not just the reachable ones")
}
