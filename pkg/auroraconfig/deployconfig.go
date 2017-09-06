package auroraconfig

import (
	"encoding/json"

	"github.com/skatteetaten/ao/pkg/serverapi"
)

type ApplicationResult struct {
	DeployId           string
	AuroraDc           serverapi.AuroraDeploymentConfig
	OpenShiftResponses []serverapi.OpenShiftResponse
}

func Response2ApplicationResults(response serverapi.Response) (applicationResults []ApplicationResult, err error) {

	applicationResults = make([]ApplicationResult, len(response.Items))
	for i := range response.Items {
		err = json.Unmarshal(response.Items[i], &applicationResults[i])
		if err != nil {
			return nil, err
		}
	}

	return applicationResults, nil
}

func ReportApplicationResuts(applicationResults []ApplicationResult) (output string) {
	var newLine = ""
	for _, applicationResult := range applicationResults {
		output += newLine + "Deploy id: " + applicationResult.DeployId + "\n"
		output += "\tApplication: " + applicationResult.AuroraDc.EnvName + "/" + applicationResult.AuroraDc.Name
		newLine = "\n"
	}
	return
}
