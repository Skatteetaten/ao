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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const SPEC_ILLEGAL = -1
const MISSING_FILE_REFERENCE = -2
const MISSING_CONFIGURATION = -3
const ILLEGAL_JSON_CONFIGURATION = -4
const CONFIGURATION_FILE_IN_ROOT = -5
const ERROR_READING_FILE = -6
const COMMUNICATION_ERROR = -7
const ILLEGAL_FILE = -8

const SPEC_IS_FILE = 1
const SPEC_IS_FOLDER = 2

var overrideFiles []string
var overrideValues []string

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   `setup folder | file [-f file 'JSON Configuration String]'`,
	Short: "Deploys an application to OpenShift based upon local configuration files",
	Long:  `Usage: When used with a `,
	Run: func(cmd *cobra.Command, args []string) {

		validateCode := validateCommand(args)
		if validateCode < 0 {
			os.Exit(validateCode)
		}

		var absolutePath string

		absolutePath, _ = filepath.Abs(args[0])

		var envFile string      // Filename for app
		var envFolder string    // Short folder name (Env)
		var folder string       // Absolute path of folder
		var parentFolder string // Absolute path of parent

		switch validateCode {
		case SPEC_IS_FILE:
			folder = filepath.Dir(absolutePath)
			envFile = filepath.Base(absolutePath)
		case SPEC_IS_FOLDER:
			folder = absolutePath
			envFile = ""
		}
		parentFolder = filepath.Dir(folder)
		envFolder = filepath.Base(folder)

		if folder == parentFolder {
			fmt.Println("Application configuration file cannot reside in root directory")
			os.Exit(CONFIGURATION_FILE_IN_ROOT)
		}

		// Initialize JSON
		var jsonStr string
		jsonStr += "{"

		// Add parameters
		jsonStr += "\"Env\":\"" + envFolder + "\""
		jsonStr += ",\"App\":\"" + envFile + "\""

		jsonStr += ",\"Files\": {"

		// Collect all JSON files in envFolder and parentFolder
		jsonStr += folder2Json(folder, envFolder+"/")
		if jsonStr != "" {
			jsonStr += ","
		}
		jsonStr += folder2Json(parentFolder, "")
		jsonStr += "}"

		// Process overrides
		jsonStr += ",\"Overrides\": {"
		jsonStr += overrides2Json(args)
		jsonStr += "}"

		// Terminate JSON
		jsonStr += "}"

		fmt.Println(string(prettyPrintJson(jsonStr)))

		callBoober(jsonStr)

	},
}

func overrides2Json(args []string) string {
	var combinedJson string

	for i := 0; i < len(overrideFiles); i++ {
		if i > 0 {
			combinedJson += ","
		}
		combinedJson += "\"" + overrideFiles[i] + "\":"
		combinedJson += args[i+1]
	}

	return combinedJson
}

func folder2Json(folder string, prefix string) string {
	var combinedJson string

	files, _ := ioutil.ReadDir(folder)
	var filesProcessed = 0
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if isLegalFileFolder(absolutePath) == SPEC_IS_FILE { // Ignore folders
			matched, _ := filepath.Match("*.json", strings.ToLower(f.Name()))
			if matched {
				filesProcessed++
				if filesProcessed > 1 {
					combinedJson += ","
				}
				combinedJson += "\"" + prefix + f.Name() + "\":"
				fileJson, err := ioutil.ReadFile(absolutePath)
				if err != nil {
					fmt.Println("Error in reading file " + absolutePath)
					os.Exit(ERROR_READING_FILE)
				}
				combinedJson += string(fileJson)
			}
		}

	}
	return combinedJson
}

func validateCommand(args []string) int {
	var errorString = ""
	var returnCode = 0

	if len(args) == 0 {
		returnCode = -1
		errorString += "Missing file/folder "
	} else {
		// Chceck argument 0 for legal file / folder
		returnCode = isLegalFileFolder(args[0])
		if returnCode < 0 {
			errorString += "Illegal file / folder: " + args[0]
			returnCode = ILLEGAL_FILE
		}

		// We have at least one argument, now there should be a correlation between the number of args
		// and the number of override (-f) flags
		if len(overrideFiles) < (len(args) - 1) {
			returnCode = MISSING_FILE_REFERENCE
			errorString += "Configuration override specified without file reference flag "
		}
		if len(overrideFiles) > (len(args) - 1) {
			returnCode = MISSING_CONFIGURATION
			errorString += "Configuration overide file reference flag specified without configuration "
		}

		// Check for legal JSON argument for each overrideFiles flag
		for i := 1; i < len(args); i++ {
			if !isLegalJson(args[i]) {
				errorString = "Illegal JSON configuration override: " + args[i] + " "
				returnCode = ILLEGAL_JSON_CONFIGURATION
			}
		}
	}

	if returnCode != 0 {
		fmt.Println(errorString)
	}
	return returnCode

}

func isLegalFileFolder(filespec string) int {
	var err error
	var absolutePath string
	var fi os.FileInfo

	absolutePath, err = filepath.Abs(filespec)
	fi, err = os.Stat(absolutePath)
	if os.IsNotExist(err) {
		return SPEC_ILLEGAL
	} else {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			return SPEC_IS_FOLDER
		case mode.IsRegular():
			return SPEC_IS_FILE
		}
	}
	return SPEC_ILLEGAL
}

func isLegalJson(jsonString string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonString), &js) == nil
}

func prettyPrintJson(jsonString string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(jsonString), "", "\t")
	if err != nil {
		return jsonString
	}
	return out.String()
}

func callBoober(combindedJson string) {
	url := "http://localhost:8080/api/setupMock/"
	var jsonStr = []byte(combindedJson)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return
	}

	req.Header.Set("OpenShiftToken", "mydirtysecret")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unable to communicate with the Boober service on " + url + ":")
		fmt.Println("Response Status: ", resp.Status)
		fmt.Println("Response Headers: ", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Response Body: ", string(body))
		os.Exit(COMMUNICATION_ERROR)
	}

}

func init() {
	RootCmd.AddCommand(setupCmd)

	// File flag, supports multiple instances of the flag
	setupCmd.Flags().StringArrayVarP(&overrideFiles, "file", "f", overrideValues, "File to override")

}
