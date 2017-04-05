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
type OpenShiftResponse struct {
	OperationType string `json:"operationType"` // CREATED eller NONE
	Payload       struct {
		Kind string `json:"kind"`
	} `json:"payload"`
}

type AuroraDc struct {
	Affiliation          string `json:"affiliation"`
	EnvName              string `json:"envName"`
	Cluster              string `json:"cluster"`
	DeploymentDescriptor struct {
		ArtifactId string `json:"artifactId"`
		GroupId    string `json:"groupId"`
		Version    string `json:"version"`
	} `json:"deployDescriptor"`
}

type ApiReturnItem struct {
	ApplicationID struct {
		EnvironmentName string `json:"environmentName"`
		ApplicationName string `json:"applicationName"`
	} `json:"applicationID"`
	AuroraDc           AuroraDc            `json:"auroraDc"`
	OpenShiftResponses []OpenShiftResponse `json:"openShiftResponses"`
}
type ApiReturn struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Items   []ApiReturnItem `json:"items"`
}

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

func GetApiSetupUrl(clusterName string, localhost bool, dryrun bool) string {
	var endpoint string = "/setup"
	if dryrun {
		endpoint = "/setup-dryrun"
	}
	return GetApiAddress(clusterName, localhost) + endpoint
}

func CallApi(combindedJson string, showConfig bool, showObjects bool, api bool, localhost bool, verbose bool,
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
		output, err = callApiInstance(combindedJson, showConfig, showObjects, verbose,
			GetApiSetupUrl("localhost", localhost, dryRun), token, dryRun, debug)
		if err != nil {
			return
		}
	} else {
		var errorString string
		var newlineErr, newlineOut string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					out, err := callApiInstance(combindedJson, showConfig, showObjects, verbose,
						GetApiSetupUrl(openshiftConfig.Clusters[i].Name, localhost, dryRun),
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

func callApiInstance(combindedJson string, showConfig bool, showObjects bool, verbose bool, url string, token string, dryRun bool, debug bool) (string, error) {
	var output string

	if showConfig {
		output += jsonutil.PrettyPrintJson(string(combindedJson))
	}

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
	bodyStr := string(body)

	if debug {
		fmt.Println("DEBUG: Response body: ")
		fmt.Println(jsonutil.PrettyPrintJson(bodyStr))
	}
	//fmt.Println("HTTP Status code: " + strconv.Itoa(resp.StatusCode))
	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusBadRequest) {
		//fmt.Println("Not StatusOK and not StatusBadRequest")
		var errorstring string
		if !strings.Contains(bodyStr, apiNotInstalledResponse) {
			errorstring = fmt.Sprintf("Internal error on %v: %v", url, bodyStr)
		}
		if verbose {
			if strings.Contains(bodyStr, apiNotInstalledResponse) {
				fmt.Println("WARN.  Boober not available")
			} else {
				fmt.Println("FAIL.  Internal error")
			}
		}
		return "", errors.New(fmt.Sprintf(errorstring))
	}

	var apiReturn ApiReturn

	if resp.StatusCode == http.StatusBadRequest {
		// We have a validation situation, give error
		if verbose {
			fmt.Println("FAIL.  Error in configuration")
		}
		return "", errors.New(fmt.Sprintf(bodyStr))
	}

	err = json.Unmarshal(body, &apiReturn)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error unmarshalling Boober return: %v\n", err.Error()))
	}

	output += ""

	if verbose {
		fmt.Println(apiReturn.Message)
	}

	var countMap map[string]int = make(map[string]int)
	for itemKey := range apiReturn.Items {
		// Loop through the applications created
		output += "Application " + apiReturn.Items[itemKey].AuroraDc.DeploymentDescriptor.GroupId + "." +
			apiReturn.Items[itemKey].AuroraDc.DeploymentDescriptor.ArtifactId + "." +
			apiReturn.Items[itemKey].AuroraDc.DeploymentDescriptor.Version +
			" deployed " +
			apiReturn.Message + " on cluster " + apiReturn.Items[itemKey].AuroraDc.Cluster + "/" +
			apiReturn.Items[itemKey].AuroraDc.Affiliation + "-" +
			apiReturn.Items[itemKey].AuroraDc.EnvName
		for osKey := range apiReturn.Items[itemKey].OpenShiftResponses {
			if apiReturn.Items[itemKey].OpenShiftResponses[osKey].OperationType == "CREATED" {
				countMap[apiReturn.Items[itemKey].OpenShiftResponses[osKey].Payload.Kind]++
			}
		}
		var space string
		var count int
		var out string
		for key := range countMap {
			out += fmt.Sprintf("%v%v: %v", space, key, countMap[key])
			space = "  "
			count++
		}
		if out != "" {
			output += " (" + out + ")"
		} else {
			output += ", no objects updated"
		}
	}

	/*if showObjects {
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
			output += fmt.Sprintf("OK.  Objects: %v  (%v)", count, out)
			fmt.Println(output)
		}

		output += jsonutil.PrettyPrintJson(string(booberReturn.OpenshiftObjects))
	}*/

	return output, nil
}
