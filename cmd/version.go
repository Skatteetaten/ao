// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
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
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/versionutil"
	"github.com/spf13/cobra"
	"os"
)

var outputFormat string
// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the aoc client",
	Long:  `Shows the version of the aoc client application`,
	Run: func(cmd *cobra.Command, args []string) {
		var output string
		var err error
		var versionStruct versionutil.VersionStruct
		versionStruct.Init()
		switch outputFormat {

		case "json":
			{
				output, err = versionStruct.Version2Json()
				output = jsonutil.PrettyPrintJson(output)
			}
		case "filename":
			{
				output, err = versionStruct.Version2Filename()
			}
		case "branch":
			{
				output, err = versionStruct.Version2Branch()
			}
		default:
			{
				output, err = versionStruct.Version2Text()
			}
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		fmt.Println(output)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	versionCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "text", "filename | json | text")
}
