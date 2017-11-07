package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type ApplicationId struct {
	Environment string `json:"environment"`
	Application string `json:"application"`
}

type DeployResult struct {
	DeployId string `json:"deployId"`
	ADS      struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Cluster   string `json:"cluster"`
	} `json:"auroraDeploymentSpec"`
	Success bool `json:"success"`
}

type DeployPayload struct {
	ApplicationIds []ApplicationId            `json:"applicationIds"`
	Overrides      map[string]json.RawMessage `json:"overrides"`
	Deploy         bool                       `json:"deploy"`
}

func NewDeployPayload(applications []string, overrides []string) (*DeployPayload, error) {
	applicationIds := createApplicationIds(applications)
	override, err := parseOverride(overrides)
	if err != nil {
		return nil, err
	}
	return &DeployPayload{
		ApplicationIds: applicationIds,
		Overrides:      override,
		Deploy:         true,
	}, nil
}

func (api *ApiClient) Deploy(deployPayload *DeployPayload) ([]DeployResult, error) {

	payload, err := json.Marshal(deployPayload)
	if err != nil {
		return nil, errors.New("failed to marshal DeployPayload")

	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)
	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var deploys []DeployResult
	err = json.Unmarshal(response.Items, &deploys)
	if err != nil {
		return nil, err
	}

	sort.Slice(deploys, func(i, j int) bool {
		return strings.Compare(deploys[i].ADS.Name, deploys[j].ADS.Name) < 1
	})

	return deploys, nil
}

func createApplicationIds(apps []string) []ApplicationId {
	applicationIds := []ApplicationId{}
	for _, app := range apps {
		envApp := strings.Split(app, "/")
		applicationIds = append(applicationIds, ApplicationId{
			Environment: envApp[0],
			Application: envApp[1],
		})
	}
	return applicationIds
}

func parseOverride(override []string) (returnMap map[string]json.RawMessage, err error) {
	returnMap = make(map[string]json.RawMessage)

	for i := 0; i < len(override); i++ {
		indexByte := strings.IndexByte(override[i], ':')
		filename := override[i][:indexByte]

		jsonOverride := override[i][indexByte+1:]
		if !IsLegalJson(jsonOverride) {
			msg := fmt.Sprintf("%s is not a valid json", jsonOverride)
			return nil, errors.New(msg)
		}
		returnMap[filename] = json.RawMessage(jsonOverride)
	}
	return returnMap, err
}

func IsLegalJson(jsonString string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonString), &js) == nil
}
