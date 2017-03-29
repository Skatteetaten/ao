package serverapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"io/ioutil"
	"net/http"
	"strings"
)

const apiNotInstalledResponse = "Application is not available"

// Structs to represent return data from the API interface
type ApiReturnObjects struct {
	Sources          json.RawMessage            `json:"sources"`
	Errors           []string                   `json:"errors"`
	Valid            bool                       `json:"valid"`
	Config           json.RawMessage            `json:"config"`
	OpenshiftObjects map[string]json.RawMessage `json:"openshiftObjects"`
}

type ApiReturn struct {
	Sources          json.RawMessage `json:"sources"`
	Errors           []string        `json:"errors"`
	Valid            bool            `json:"valid"`
	Config           json.RawMessage `json:"config"`
	OpenshiftObjects json.RawMessage `json:"openshiftObjects"`
}

func GetApiAddress(clusterName string, localhost bool) (apiAddress string) {
	if localhost {
		apiAddress = "http://localhost:8080"
	} else {
		apiAddress = "http://boober-mfp-boober." + clusterName + ".paas.skead.no"
	}
	return
}

// Check for valid login, that is we have a configuration with at least one reachable cluster
func ValidateLogin(openshiftConfig *openshift.OpenshiftConfig) (output bool) {
	var openshiftCluster *openshift.OpenshiftCluster
	openshiftCluster, _ = openshiftConfig.GetApiCluster()
	if openshiftCluster != nil {
		if !openshiftCluster.HasValidToken() {
			return false
		}
	}
	return true
}

func GetApiSetupUrl(clusterName string, localhost bool) string {
	return GetApiAddress(clusterName, localhost) + "/setup"
}

func CallApi(combindedJson string, showConfig bool, showObjects bool, api bool, localhost bool, verbose bool,
	openshiftConfig *openshift.OpenshiftConfig) (output string, err error) {
	//var openshiftConfig *openshift.OpenshiftConfig
	var apiCluster *openshift.OpenshiftCluster

	if localhost {
		var token string = ""
		apiCluster, err = openshiftConfig.GetApiCluster()
		if apiCluster != nil {
			token = apiCluster.Token
		}
		output, err = callApiInstance(combindedJson, showConfig, showObjects, verbose,
			GetApiSetupUrl("localhost", localhost), token)
		if err != nil {
			return
		}
	} else {
		var errorString string
		var newline string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					out, err := callApiInstance(combindedJson, showConfig, showObjects, verbose,
						GetApiSetupUrl(openshiftConfig.Clusters[i].Name, localhost),
						openshiftConfig.Clusters[i].Token)
					if err == nil {
						output += fmt.Sprintf("%v\n", out)
					} else {
						if err.Error() != "" {
							errorString += newline + err.Error()
							newline = "\n"
						}
					}
				}
			}
		}
		if errorString != "" {
			return output, errors.New(errorString)
		}
	}
	return output, nil
}

func callApiInstance(combindedJson string, showConfig bool, showObjects bool, verbose bool, url string, token string) (string, error) {
	var output string

	if verbose {
		fmt.Print("Sending config to Boober at " + url + "... ")
	}

	var jsonStr = []byte(combindedJson)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Internal error in NewRequest: %v", err))
	}

	req.Header.Set("Authentication", "Bearer: "+token)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("FAIL.  Error connecting to Boober service")
		}
		return "", errors.New(fmt.Sprintf("Error connecting to the Boober service on %v: %v", url, err))
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		var errorstring string
		if strings.Contains(bodyStr, apiNotInstalledResponse) {
			errorstring = fmt.Sprintf("Boober not available on %v", url)
		} else {
			errorstring = fmt.Sprintf("Internal error")
		}
		if verbose {
			if strings.Contains(bodyStr, apiNotInstalledResponse) {
				fmt.Println("FAIL.  Boober not available")
			} else {
				fmt.Println("FAIL.  Internal error")
			}
		}
		return "", errors.New(fmt.Sprintf(errorstring))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	// Check return for error
	var booberReturn ApiReturn
	err = json.Unmarshal(body, &booberReturn)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error unmarshalling Boober return: %v\n", err.Error()))
	}

	for _, message := range booberReturn.Errors {
		fmt.Println("DEBUG: Error from Boober:  " + message)
	}
	if !(booberReturn.Valid) {
		fmt.Println("Error in configuration: ")
		for _, message := range booberReturn.Errors {
			fmt.Println("  " + message)
		}
	} else {
		if verbose {
			var apiReturnObjects ApiReturnObjects
			err = json.Unmarshal(body, &apiReturnObjects)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Error unmarshalling Boober return: %v\n", err.Error()))
			}
			var countMap map[string]int = make(map[string]int)
			for key := range apiReturnObjects.OpenshiftObjects {
				countMap[key]++
			}

			var space string
			var count int
			var out string
			for key := range countMap {
				out += fmt.Sprintf("%v%v: %v", space, key, countMap[key])
				space = "  "
				count++
			}
			if count > 0 {
				output := fmt.Sprintf("OK.  Objects: %v  (%v)", count, out)
				fmt.Println(output)
			}
		}
	}

	if showConfig {
		output += jsonutil.PrettyPrintJson(string(booberReturn.Config))
	}

	if showObjects {
		output += jsonutil.PrettyPrintJson(string(booberReturn.OpenshiftObjects))
	}

	return output, nil
}
