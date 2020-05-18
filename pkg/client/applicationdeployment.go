package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/skatteetaten/ao/pkg/deploymentspec"
)

// ApplicationDeploymentClient is a client for deploying an application via an external service
type ApplicationDeploymentClient interface {
	Doer
	Deploy(deployPayload *DeployPayload) (*DeployResults, error)
	Delete(deletePayload *DeletePayload) (*DeleteResults, error)
	Exists(existPayload *ExistsPayload) (*ExistsResults, error)
	GetApplyResult(deployID string) (string, error)
}

type (
	// ApplicationDeploymentRef specifies an application deployment reference
	ApplicationDeploymentRef struct {
		Environment string `json:"environment"`
		Application string `json:"application"`
	}
	// ApplicationRef references an application in a namespace
	ApplicationRef struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	// DeployResults holds the results of deployments
	DeployResults struct {
		Message string
		Success bool
		Results []DeployResult
	}

	// DeployResult holds the result of a deployment
	DeployResult struct {
		DeployID       string                        `json:"deployId"`
		DeploymentSpec deploymentspec.DeploymentSpec `json:"deploymentSpec"`
		Success        bool                          `json:"success"`
		Ignored        bool                          `json:"ignored"`
		Reason         string                        `json:"reason"`
		Warnings       []string                      `json:"warnings"`
	}

	// DeployPayload holds the payload of a deployment
	DeployPayload struct {
		ApplicationDeploymentRefs []ApplicationDeploymentRef `json:"applicationDeploymentRefs"`
		Overrides                 map[string]string          `json:"overrides"`
		Deploy                    bool                       `json:"deploy"`
	}

	// DeleteResults hold the results from deleting applications
	DeleteResults struct {
		Message string
		Success bool
		Results []DeleteResult
	}

	// DeleteResult holds the results from deleting an application
	DeleteResult struct {
		ApplicationRef ApplicationRef `json:"applicationRef"`
		Success        bool           `json:"success"`
		Reason         string         `json:"reason"`
	}

	// DeletePayload holds references to applications that are to be deleted
	DeletePayload struct {
		ApplicationRefs []ApplicationRef `json:"applicationRefs"`
	}

	// ExistsResults hold information on which applications that exist
	ExistsResults struct {
		Message string
		Success bool
		Results []ExistsResult
	}

	// ExistsResult hold information on whether an application exists
	ExistsResult struct {
		ApplicationRef ApplicationRef `json:"applicationRef"`
		Exists         bool           `json:"exists"`
		Success        bool           `json:"success"`
		Message        string         `json:"message"`
	}

	// ExistsPayload is the payload of an enquiry action for application existence
	ExistsPayload struct {
		ApplicationDeploymentRefs []ApplicationDeploymentRef `json:"adr"`
	}
)

// NewApplicationDeploymentRef creates an ApplicationDeploymentRef
func NewApplicationDeploymentRef(name string) *ApplicationDeploymentRef {
	slice := strings.Split(name, "/")
	return &ApplicationDeploymentRef{
		Environment: slice[0],
		Application: slice[1],
	}
}

// NewApplicationRef creates an ApplicationRef
func NewApplicationRef(namespace, name string) *ApplicationRef {
	return &ApplicationRef{
		Namespace: namespace,
		Name:      name,
	}
}

// NewDeployPayload creates a DeployPayload
func NewDeployPayload(applications []string, overrides map[string]string) *DeployPayload {
	applicationDeploymentRefs := createApplicationDeploymentRefs(applications)
	return &DeployPayload{
		ApplicationDeploymentRefs: applicationDeploymentRefs,
		Overrides:                 overrides,
		Deploy:                    true,
	}
}

// NewDeletePayload create a DeletePayload
func NewDeletePayload(applicationRefs []ApplicationRef) *DeletePayload {
	return &DeletePayload{
		ApplicationRefs: applicationRefs,
	}
}

// NewExistsPayload create an ExistsPayload
func NewExistsPayload(applications []string) *ExistsPayload {
	applicationDeploymentRefs := createApplicationDeploymentRefs(applications)
	return &ExistsPayload{
		ApplicationDeploymentRefs: applicationDeploymentRefs,
	}
}

// Deploy deploys application(s) as specified in a DeployPayload
func (api *APIClient) Deploy(deployPayload *DeployPayload) (*DeployResults, error) {
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

// Delete deletes application(s) as specified in a DeletePayload
func (api *APIClient) Delete(deletePayload *DeletePayload) (*DeleteResults, error) {
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

// Exists queries the existence of application(s) as specified in an ExistsPayload
func (api *APIClient) Exists(existsPayload *ExistsPayload) (*ExistsResults, error) {
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

func createApplicationDeploymentRefs(apps []string) []ApplicationDeploymentRef {
	var applicationDeploymentRefs []ApplicationDeploymentRef
	for _, app := range apps {
		applicationDeploymentRefs = append(applicationDeploymentRefs, *NewApplicationDeploymentRef(app))
	}
	return applicationDeploymentRefs
}
