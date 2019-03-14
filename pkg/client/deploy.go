package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type DeployClient interface {
	Doer
	Deploy(deployPayload *DeployPayload) (*DeployResults, error)
	GetApplyResult(deployId string) (string, error)
}

type (
	applicationDeploymentRef struct {
		Environment string `json:"environment"`
		Application string `json:"application"`
	}

	AuroraConfigFieldSource struct {
		Value interface{} `json:"value"`
	}

	DeploymentSpec map[string]AuroraConfigFieldSource

	DeployResults struct {
		Message string
		Success bool
		Results []DeployResult
	}

	DeployResult struct {
		DeployId       string         `json:"deployId"`
		DeploymentSpec DeploymentSpec `json:"deploymentSpec"`
		Success        bool           `json:"success"`
		Ignored        bool           `json:"ignored"`
		Reason         string         `json:"reason"`
	}

	DeployPayload struct {
		ApplicationDeploymentRefs []applicationDeploymentRef `json:"applicationDeploymentRefs"`
		Overrides                 map[string]string          `json:"overrides"`
		Deploy                    bool                       `json:"deploy"`
	}
)

func (spec DeploymentSpec) Get(name string) interface{} {
	field, ok := spec[name]
	if !ok {
		return nil
	}
	return field.Value
}

func (spec DeploymentSpec) GetString(name string) string {
	if value := spec.Get(name); value != nil {
		return value.(string)
	} else {
		return ""
	}
}

func (spec DeploymentSpec) Cluster() string {
	return spec.GetString("cluster")
}

func (spec DeploymentSpec) Environment() string {
	return spec.GetString("envName")
}

func (spec DeploymentSpec) Name() string {
	return spec.GetString("name")
}

func (spec DeploymentSpec) Version() string {
	return spec.GetString("version")
}

func NewAuroraConfigFieldSource(value interface{}) AuroraConfigFieldSource {
	return AuroraConfigFieldSource{
		Value: value,
	}
}

func NewApplicationDeploymentRef(name string) *applicationDeploymentRef {
	slice := strings.Split(name, "/")
	return &applicationDeploymentRef{
		Environment: slice[0],
		Application: slice[1],
	}
}

func NewDeployPayload(applications []string, overrides map[string]string) *DeployPayload {
	applicationDeploymentRefs := createApplicationDeploymentRefs(applications)
	return &DeployPayload{
		ApplicationDeploymentRefs: applicationDeploymentRefs,
		Overrides:                 overrides,
		Deploy:                    true,
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

	if !response.Success {
		for _, deploy := range deploys.Results {
			// Hack-ish solution since validation errors and deploy errors have
			// different payload. TODO: Fix error response from Boober.
			if deploy.DeploymentSpec.Name() == "" {
				return nil, response.Error()
			}
		}
	}

	deploys.Message = response.Message
	if response.Success {
		deploys.Success = true
	}

	return &deploys, nil
}

func createApplicationDeploymentRefs(apps []string) []applicationDeploymentRef {
	var applicationDeploymentRefs []applicationDeploymentRef
	for _, app := range apps {
		applicationDeploymentRefs = append(applicationDeploymentRefs, *NewApplicationDeploymentRef(app))
	}
	return applicationDeploymentRefs
}
