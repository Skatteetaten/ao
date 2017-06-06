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
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/versionutil"
	"github.com/spf13/cobra"
	"os"
)

var majorVersion = "5"
var minorVersion = "0"
var buildstamp = ""
var githash = ""
var buildnumber = ""
var branch = ""

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the aoc client",
	Long:  `Shows the version of the aoc client application`,
	Run: func(cmd *cobra.Command, args []string) {
		var output string
		var err error
		switch outputFormat {

		case "json":
			{
				output, err = versionutil.Version2Json(majorVersion, minorVersion, buildnumber, githash, branch, buildstamp)
				output = jsonutil.PrettyPrintJson(output)
			}
		case "filename":
			{
				output = "aoc_" + majorVersion + "." + minorVersion + "." + buildnumber
			}
		case "branch":
			{
				output = branch
			}
		default:
			{
				output, err = versionutil.Version2Text(majorVersion, minorVersion, buildnumber, githash, branch, buildstamp)
			}
		}
		if err != nil {
			fmt.Print(err.Error())
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
