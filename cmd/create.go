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
	"github.com/skatteetaten/aoc/pkg/createcmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create vault <vaultname> | secret <vaultname> <secretname>",
	Short: "Creates a vault or a secret in a vault",
	Long:  `Creates a vault or a secret in a vault.`,
	Run: func(cmd *cobra.Command, args []string) {
		var createcmdObject createcmd.CreatecmdClass
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
