package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/createcmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var createCmd = &cobra.Command{
	Use:   "create vault <vaultname> | secret <vaultname> <secretname>",
	Short: "Creates a vault or a secret in a vault",
	Long:  `Creates a vault or a secret in a vault.`,
	Run: func(cmd *cobra.Command, args []string) {
		var createcmdObject createcmd.CreatecmdClass

		allClusters, _ := cmd.Flags().GetBool("all")
		output, err := createcmdObject.CreateObject(args, &persistentOptions, allClusters)
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
	RootCmd.AddCommand(createCmd)

	getClusterCmd.Flags().BoolP("all",
		"a", false, "Show all clusters, not just the reachable ones")
}
