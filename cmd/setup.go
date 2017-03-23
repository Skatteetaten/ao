// Copyright © 2017 Norwegian Tax Authority
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
	"github.com/skatteetaten/aoc/pkg/boober"
	"github.com/spf13/cobra"
)

// Cobra Flag variables
var overrideFiles []string
var overrideValues []string
var dryRun bool
var showConfig bool
var localhost bool
var showObjects bool

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   `setup folder | file [-f file 'JSON Configuration String']`,
	Short: "Deploys an application to OpenShift based upon local configuration files",
	Long: `When used with a .json file as an argument, it will deploy the application referred to in the
file merged with about.json in the same folder, and about.json and aos-features.json in the parent folder`,
	Run: func(cmd *cobra.Command, args []string) {
		boober.ExecuteSetup(args, dryRun, showConfig, localhost, overrideFiles)
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)

	// File flag, supports multiple instances of the flag
	setupCmd.Flags().StringArrayVarP(&overrideFiles, "file",
		"f", overrideValues, "File to override")
	setupCmd.Flags().BoolVarP(&dryRun, "dryrun",
		"d", false,
		"Do not perform a setup, just collect and print the configuration files")
	setupCmd.Flags().BoolVarP(&showConfig, "showconfig",
		"s", false, "Print merged config from Boober to standard out")
	setupCmd.Flags().BoolVarP(&showObjects, "showobjects",
		"o", false, "Print object definitions from Boober to standard out")
	setupCmd.Flags().BoolVarP(&localhost, "localhost",
		"l", false, "Send setup to Boober on localhost")
}
