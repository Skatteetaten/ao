package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type (
	applicationId struct {
		Environment string `json:"environment"`
		Application string `json:"application"`
	}

	DeployResults struct {
		Message string
		Success bool
		Results []DeployResult
	}

	DeployResult struct {
		DeployId string `json:"deployId"`
		ADS      struct {
			Name    string `json:"name"`
			Cluster string `json:"cluster"`
			Deploy  struct {
				Version string `json:"version"`
			} `json:"deploy"`
			Environment struct {
				Namespace string `json:"namespace"`
			} `json:"environment"`
		} `json:"auroraDeploymentSpec"`
		Success bool   `json:"success"`
		Ignored bool   `json:"ignored"`
		Reason  string `json:"reason"`
	}

	DeployPayload struct {
		ApplicationIds []applicationId            `json:"applicationIds"`
		Overrides      map[string]json.RawMessage `json:"overrides"`
		Deploy         bool                       `json:"deploy"`
	}
)

func NewDeployPayload(applications []string, overrides map[string]json.RawMessage) *DeployPayload {
	applicationIds := createApplicationIds(applications)
	return &DeployPayload{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
		Deploy:         true,
	}
}

func (api *ApiClient) Deploy(deployPayload *DeployPayload) (*DeployResults, error) {

	payload, err := json.Marshal(deployPayload)
	if err != nil {
		return nil, errors.New("failed to marshal DeployPayload")

	}

	endpoint := fmt.Sprintf("/apply/%s", api.Affiliation)
	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var deploys DeployResults
	err = json.Unmarshal(response.Items, &deploys.Results)
	if err != nil {
		return nil, err
	}

	deploys.Message = response.Message
	if response.Success {
		deploys.Success = true
	}

	return &deploys, nil
}

func createApplicationIds(apps []string) []applicationId {
	var applicationIds []applicationId
	for _, app := range apps {
		envApp := strings.Split(app, "/")
		applicationIds = append(applicationIds, applicationId{
			Environment: envApp[0],
			Application: envApp[1],
		})
	}
	return applicationIds
}
