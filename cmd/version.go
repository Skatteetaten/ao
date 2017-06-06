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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type VersionStruct struct {
	MajorVersion string
	MinorVersion string
	BuildNumber  string
	Version      string
	Branch       string
}

var version = "5"
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
		if outputFormat == "json" {
			output, err = version2Json()
		} else {
			output, err = version2Text()
		}
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(-1)
		}
		fmt.Println(output)
	},
}

func version2Text() (output string, err error) {
	if buildnumber != "" {
		version = version + "." + buildnumber
	}
	output = "Aurora OC version " + majorVersion + "." + minorVersion + "." + buildnumber
	if githash != "" {
		output += "\nBranch: " + branch + " (" + githash + ")"
	}
	if buildstamp != "" {
		output += "\nBuild Time: " + buildstamp
	}
	return
}

func version2Json() (output string, err error) {

	var versionStruct VersionStruct
	versionStruct.MajorVersion = majorVersion
	versionStruct.MinorVersion = minorVersion
	versionStruct.BuildNumber = buildnumber
	versionStruct.Version = majorVersion + "." + minorVersion + "." + buildnumber
	versionStruct.Branch = branch
	outputBytes, err := json.Marshal(versionStruct)
	if err != nil {
		return
	}
	output = string(outputBytes)
	return
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

	versionCmd.Flags().StringVarP(&outputFormat, "output-format", "o", "text", "text | json")
}
