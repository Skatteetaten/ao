// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/pingcmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var pingPort string
var pingCluster string

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Checks for open connectivity from all nodes in the cluster to a specific ip address and port. ",
	Long: `Invokes the network debug service in the Aurora Console
to ping the specified address and port from each node.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pingObject pingcmd.PingcmdClass
		output, err := pingObject.PingAddress(args, pingPort, pingCluster, &persistentOptions)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pingCmd.Flags().StringVarP(&pingPort, "port", "p", "80", "Port to ping")
	pingCmd.Flags().StringVarP(&pingCluster, "cluster", "c", "", "OpenShift source cluster")
}
