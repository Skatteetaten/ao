package serverapi_v2

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"io/ioutil"
	"net/http"
	"strings"
)

const apiNotInstalledResponse = "Application is not available"

func GetApiAddress(clusterName string, localhost bool) (apiAddress string) {
	if localhost {
		apiAddress = "http://localhost:8080"
	} else {
		apiAddress = "http://boober-aos-bas-dev." + clusterName + ".paas.skead.no"
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

func GetApiSetupUrl(clusterName string, apiEndpont string, localhost bool, dryrun bool) string {
	return GetApiAddress(clusterName, localhost) + apiEndpont
}

func CallApi(apiEndpoint string, combindedJson string, showConfig bool, showObjects bool, api bool, localhost bool, verbose bool,
	openshiftConfig *openshift.OpenshiftConfig, dryRun bool, debug bool) (output string, err error) {
	//var openshiftConfig *openshift.OpenshiftConfig
	var apiCluster *openshift.OpenshiftCluster

	if localhost {
		var token string = ""
		apiCluster, err = openshiftConfig.GetApiCluster()
		if apiCluster != nil {
			token = apiCluster.Token
			if debug {
				fmt.Println("DEBUG: Token to Localhost: " + token)
			}
		}
		output, err = callApiInstance(combindedJson, verbose,
			GetApiSetupUrl("localhost", apiEndpoint, localhost, dryRun), token, dryRun, debug)
		if err != nil {
			return
		}
	} else {
		var errorString string
		var newlineErr, newlineOut string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					out, err := callApiInstance(combindedJson, verbose,
						GetApiSetupUrl(openshiftConfig.Clusters[i].Name, apiEndpoint, localhost, dryRun),
						openshiftConfig.Clusters[i].Token, dryRun, debug)
					if err == nil {
						if out != "" {
							output += fmt.Sprintf("%v %v", out, newlineOut)
							newlineOut = "\n"
						}
					} else {
						if err.Error() != "" {
							errorString += newlineErr + err.Error()
							newlineErr = "\n"
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

func callApiInstance(combindedJson string, verbose bool, url string, token string, dryRun bool, debug bool) (output string, err error) {

	if verbose {
		fmt.Print("Sending config to Boober at " + url + "... ")
	}

	var jsonStr = []byte(combindedJson)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Internal error in NewRequest: %v", err))
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("dryrun", fmt.Sprintf("%v", dryRun))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("FAIL.  Error connecting to Boober service")
		}
		return "", errors.New(fmt.Sprintf("Error connecting to the Boober service on %v: %v", url, err))
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	output = string(body)

	if debug {
		fmt.Println("DEBUG: Response body: ")
		fmt.Println(jsonutil.PrettyPrintJson(output))
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusBadRequest) {

		var errorstring string
		if !strings.Contains(output, apiNotInstalledResponse) {
			errorstring = fmt.Sprintf("Internal error on %v: %v", url, output)
		}
		if verbose {
			if strings.Contains(output, apiNotInstalledResponse) {
				fmt.Println("WARN.  Boober not available")
			} else {
				fmt.Println("FAIL.  Internal error")
			}
		}
		err = errors.New(fmt.Sprintf(errorstring))
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		// We have a validation situation, give error
		if verbose {
			fmt.Println("FAIL.  Error in configuration")
		}

		err = errors.New(fmt.Sprintf(output))

		return
	}

	return
}
