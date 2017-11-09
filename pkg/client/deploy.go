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

func NewDeployPayload(applications []string, overrides map[string]json.RawMessage) (*DeployPayload, error) {
	applicationIds := createApplicationIds(applications)
	return &DeployPayload{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
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
	// If deploy fail, the DeployResult will contain errors so we can't check if the
	// Response is successfully or not
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
