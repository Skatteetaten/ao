package serverapi_v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"io/ioutil"
	"net/http"
	"strings"
)

// Structs to represent return data from the API interface

type ApplicationId struct {
	EnvironmentName string `json:"environmentName"`
	ApplicationName string `json:"applicationName"`
}

type OpenShiftResponse struct {
	Kind          string `json:"kind"`
	OperationType string `json:"operationType"` // CREATED, UPDATE eller NONE
	Payload       struct {
		Kind string `json:"payload"`
	} `json:"payload"`
	ResponseBody json.RawMessage `json:"responseBody"`
}

/*type AuroraDc struct {
	Affiliation          string `json:"affiliation"`
	EnvName              string `json:"envName"`
	Cluster              string `json:"cluster"`
	DeploymentDescriptor struct {
		ArtifactId string `json:"artifactId"`
		GroupId    string `json:"groupId"`
		Version    string `json:"version"`
	} `json:"deployDescriptor"`
}*/

type DeploymentDescriptor struct {
	TemplateFile string            `json:"templateFile"`
	Template     string            `json:"template"`
	Parameters   map[string]string `json:"parameters"`
}

type AuroraDeploymentConfig struct {
	SchemaVersion        string               `json:"schemaVersion"`
	Affiliation          string               `json:"affiliation"`
	Cluster              string               `json:"cluster"`
	Type                 string               `json:"type"`
	Name                 string               `json:"name"`
	EnvName              string               `json:"envName"`
	Groups               []string             `json:"groups"`
	Users                []string             `json:"users"`
	Replicas             int                  `json:"replicas"`
	Secrets              map[string]string    `json:"secrets"`
	Config               map[string]string    `json:"config"`
	GroupId              string               `json:"groupId"`
	ArtifactId           string               `json:"artifactId"`
	Version              string               `json:"version"`
	Route                bool                 `json:"route"`
	DeploymentStrategy   string               `json:"deploymentStrategy"`
	DeploymentDescriptor DeploymentDescriptor `json:"deploymentDescriptor"`
}

type ApplicationResult struct {
	ApplicationId     ApplicationId          `json:"applicationId"`
	AuroraDc          AuroraDeploymentConfig `json:"auroraDc"`
	OpenShiftResponse OpenShiftResponse      `json:"openShiftResponse"`
}

type Response struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Items   []json.RawMessage `json:"items"`
	Count   int               `json:"count"`
}

type ResponeItemError struct {
	ApplicationId ApplicationId `json:"applicationId"`
	Messages      []string      `json:"messages"`
}

const apiNotInstalledResponse = "Application is not available"
const localhostAddress = "localhost"
const localhostPort = "8080"

func ParseResponse(responseString string) (response Response, err error) {
	var responseData []byte
	responseData = []byte(responseString)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return
	}

	return
}

func ResponseItems2ApplicationResults(response Response) (applicationResults []ApplicationResult, err error) {
	applicationResults = make([]ApplicationResult, len(response.Items))

	for item := range response.Items {
		err = json.Unmarshal([]byte(response.Items[item]), &applicationResults[item])
	}
	return
}

func ApplicationResult2MessageString(applicationResult ApplicationResult) (output string, err error) {

	output += "\n" +
		//applicationResult.ApplicationId.ApplicationName +
		applicationResult.AuroraDc.GroupId + "/" + applicationResult.AuroraDc.ArtifactId + "-" + applicationResult.AuroraDc.Version +
		" deployed in " + applicationResult.AuroraDc.Cluster + "/" + applicationResult.AuroraDc.EnvName
	return
}

func ResponsItems2MessageString(response Response) (output string, err error) {
	if response.Message != "" {
		output = response.Message + ": "
	}

	for item := range response.Items {
		var responseItemError ResponeItemError
		err = json.Unmarshal([]byte(response.Items[item]), &responseItemError)
		if err != nil {
			return
		}
		output = output + "\n\t" + responseItemError.ApplicationId.EnvironmentName + "/" + responseItemError.ApplicationId.ApplicationName + ":"

		for message := range responseItemError.Messages {
			output = output + "\n\t\t" + responseItemError.Messages[message]
		}
	}
	return
}

func GetApiAddress(clusterName string, localhost bool) (apiAddress string) {
	if localhost {
		apiAddress = "http://" + localhostAddress + ":" + localhostPort
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
	openshiftConfig *openshift.OpenshiftConfig, dryRun bool, debug bool) (outputMap map[string]string, err error) {
	//var openshiftConfig *openshift.OpenshiftConfig
	var apiCluster *openshift.OpenshiftCluster

	outputMap = make(map[string]string)
	if localhost {
		var token string = ""
		apiCluster, err = openshiftConfig.GetApiCluster()
		if apiCluster != nil {
			token = apiCluster.Token
		}
		output, err := callApiInstance(combindedJson, verbose,
			GetApiSetupUrl(localhostAddress, apiEndpoint, localhost, dryRun), token, dryRun, debug)
		outputMap[openshiftConfig.Clusters[0].Name] = output
		if err != nil {
			return outputMap, err
		}
		//outputMap["localhost"] = output
	} else {
		var errorString string
		var newlineErr string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					output, err := callApiInstance(combindedJson, verbose,
						GetApiSetupUrl(openshiftConfig.Clusters[i].Name, apiEndpoint, localhost, dryRun),
						openshiftConfig.Clusters[i].Token, dryRun, debug)
					if output != "" {
						//fmt.Println("Debug: Setting outputMap: " + openshiftConfig.Clusters[i].Name + ":" + output)
						outputMap[openshiftConfig.Clusters[i].Name] = output

						if err != nil {
							errorString += newlineErr + err.Error()
							newlineErr = "\n"
						}
					}
				}
			}
		}
		if errorString != "" {
			err = errors.New(errorString)
			return
		}
	}
	return
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
		fmt.Println("Debug: Error in client.Do")
		return "", errors.New(fmt.Sprintf("Error connecting to the Boober service on %v: %v", url, err))
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	output = string(body)
	//fmt.Println("Debug output: " + jsonutil.PrettyPrintJson(output))

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

	if verbose {
		fmt.Print("OK")
	}
	return
}
