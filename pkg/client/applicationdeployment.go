package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/skatteetaten/ao/pkg/deploymentspec"
)

type ApplicationDeploymentClient interface {
	Doer
	Deploy(deployPayload *DeployPayload) (*DeployResults, error)
	Delete(deletePayload *DeletePayload) (*DeleteResults, error)
	Exists(existPayload *ExistsPayload) (*ExistsResults, error)
	GetApplyResult(deployId string) (string, error)
}

type (
	applicationDeploymentRef struct {
		Environment string `json:"environment"`
		Application string `json:"application"`
	}

	ApplicationRef struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}

	DeployResults struct {
		Message string
		Success bool
		Results []DeployResult
	}

	DeployResult struct {
		DeployId       string                        `json:"deployId"`
		DeploymentSpec deploymentspec.DeploymentSpec `json:"deploymentSpec"`
		Success        bool                          `json:"success"`
		Ignored        bool                          `json:"ignored"`
		Reason         string                        `json:"reason"`
		Warnings       []string                      `json:"warnings"`
	}

	DeployPayload struct {
		ApplicationDeploymentRefs []applicationDeploymentRef `json:"applicationDeploymentRefs"`
		Overrides                 map[string]string          `json:"overrides"`
		Deploy                    bool                       `json:"deploy"`
	}

	DeleteResults struct {
		Message string
		Success bool
		Results []DeleteResult
	}

	DeleteResult struct {
		ApplicationRef ApplicationRef `json:"applicationRef"`
		Success        bool           `json:"success"`
		Reason         string         `json:"reason"`
	}

	DeletePayload struct {
		ApplicationRefs []ApplicationRef `json:"applicationRefs"`
	}

	ExistsResults struct {
		Message string
		Success bool
		Results []ExistsResult
	}

	ExistsResult struct {
		ApplicationRef ApplicationRef `json:"applicationRef"`
		Exists         bool           `json:"exists"`
		Success        bool           `json:"success"`
		Message        string         `json:"message"`
	}

	ExistsPayload struct {
		ApplicationDeploymentRefs []applicationDeploymentRef `json:"adr"`
	}
)

func NewApplicationDeploymentRef(name string) *applicationDeploymentRef {
	slice := strings.Split(name, "/")
	return &applicationDeploymentRef{
		Environment: slice[0],
		Application: slice[1],
	}
}

func NewApplicationRef(namespace, name string) *ApplicationRef {
	return &ApplicationRef{
		Namespace: namespace,
		Name:      name,
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

func NewDeletePayload(applicationRefs []ApplicationRef) *DeletePayload {
	return &DeletePayload{
		ApplicationRefs: applicationRefs,
	}
}

func NewExistsPayload(applications []string) *ExistsPayload {
	applicationDeploymentRefs := createApplicationDeploymentRefs(applications)
	return &ExistsPayload{
		ApplicationDeploymentRefs: applicationDeploymentRefs,
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
			if deploy.DeploymentSpec.Name() == "-" {
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

func (api *ApiClient) Delete(deletePayload *DeletePayload) (*DeleteResults, error) {
	payload, err := json.Marshal(deletePayload)
	if err != nil {
		return nil, errors.New("Failed to marshal DeletePayload")
	}

	endpoint := fmt.Sprintf("/applicationdeployment/delete")
	response, err := api.Do(http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var deleteResults DeleteResults
	err = json.Unmarshal(response.Items, &deleteResults.Results)
	if err != nil {
		return nil, err
	}

	deleteResults.Message = response.Message
	if response.Success {
		deleteResults.Success = true
	}

	return &deleteResults, nil
}

func (api *ApiClient) Exists(existsPayload *ExistsPayload) (*ExistsResults, error) {
	payload, err := json.Marshal(existsPayload)
	if err != nil {
		return nil, errors.New("Failed to marshal ExistsPayload")
	}

	endpoint := fmt.Sprintf("/applicationdeployment/%s", api.Affiliation)
	response, err := api.Do(http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, err
	}

	var existsResults ExistsResults
	err = json.Unmarshal(response.Items, &existsResults.Results)
	if err != nil {
		return nil, err
	}

	existsResults.Message = response.Message
	if response.Success {
		existsResults.Success = true
	}

	return &existsResults, nil
}

func createApplicationDeploymentRefs(apps []string) []applicationDeploymentRef {
	var applicationDeploymentRefs []applicationDeploymentRef
	for _, app := range apps {
		applicationDeploymentRefs = append(applicationDeploymentRefs, *NewApplicationDeploymentRef(app))
	}
	return applicationDeploymentRefs
}
