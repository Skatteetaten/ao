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
	"github.com/skatteetaten/aoc/pkg/updatecmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var simulate bool
var forceVersion string
var forceUpdate bool

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates for the aoc client, and downloads the update if available.",
	Long:  `Available updates are searced for using a service in the OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := updatecmd.UpdateSelf(args, simulate, forceVersion, forceUpdate)
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
	RootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	updateCmd.Flags().BoolVarP(&simulate, "simulate", "s", false,
		"No action, just checks for avaliable update.")
	updateCmd.Flags().StringVarP(&forceVersion, "force-version", "", "", "Force download of a specific version")
	updateCmd.Flags().BoolVarP(&forceUpdate, "force-update", "", false, "Force update even if no new version")
}
