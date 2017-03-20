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
	"github.com/spf13/cobra"
	"net/http"
	//"net/url"
	"log"
	//"encoding/json"
	//"bytes"
	"bytes"
	"io/ioutil"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup <env>/<app>",
	Short: "Deploys an application to OpenShift based upon local configuration files",
	Long:  `Sends all your configuration files to Boober.  .`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")

		url := fmt.Sprintf("http://localhost:8080/api/setupMock/utv/myapp")
		var jsonStr = []byte(`{"Files": {"File1.json": {"Config1" : "foo", "Config2" : "bar"},"File2.json": {"Config21" : "xhost", "Config22" : "yhost"}}, "Overrides": {"OverideFile1.json": {"OverrideConfig1" : "foobar", "OverrideConfig2" : "badekar"}}}`)
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr)) // bytes.NewBuffer(jsonStr)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			log.Fatal("NewRequest: ", err)
			return
		}

		req.Header.Set("token", "mydirtysecret")
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Do: ", err)
			return
		}

		defer resp.Body.Close()
		fmt.Println("Response Status: ", resp.Status)
		fmt.Println("Response Headers: ", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Response Body: ", string(body))

		fmt.Println("setup finished")
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
