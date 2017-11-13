package cmd

import (
	"fmt"
	pkgPingCmd "github.com/skatteetaten/ao/pkg/pingcmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Checks for open connectivity from all nodes in the cluster to a specific ip address and port. ",
	Long: `Invokes the network debug service in the Aurora Console
to ping the specified address and port from each node.`,
	Annotations: map[string]string{"type": "util"},
	Run: func(cmd *cobra.Command, args []string) {

		pingPort, _ := cmd.Flags().GetString("port")
		pingCluster, _ := cmd.Flags().GetString("cluster")

		output, err := pkgPingCmd.PingAddress(args, pingPort, pingCluster, oldConfig)
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
	RootCmd.AddCommand(pingCmd)

	pingCmd.Flags().StringP("port", "p", "80", "Port to ping")
	pingCmd.Flags().StringP("cluster", "c", "", "OpenShift source cluster")
}
