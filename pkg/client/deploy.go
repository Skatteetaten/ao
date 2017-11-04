package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ApplicationId struct {
	Environment string `json:"environment"`
	Application string `json:"application"`
}

type deployResult struct {
	DeployId string `json:"deployId"`
	ADS      struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Cluster   string `json:"cluster"`
	} `json:"auroraDeploymentSpec"`
	Success bool `json:"success"`
}

type applyPayload struct {
	ApplicationIds []ApplicationId            `json:"applicationIds"`
	Overrides      map[string]json.RawMessage `json:"overrides"`
	Deploy         bool                       `json:"deploy"`
}

func (api *ApiClient) Deploy(applications []string, overrides map[string]json.RawMessage) error {

	applicationIds := createApplicationIds(applications)
	applyPayload := &applyPayload{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
		Deploy:         true,
	}

	payload, err := json.Marshal(applyPayload)
	if err != nil {
		fmt.Println("Failed to marshal ApplyPayload")
		return nil
	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)
	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	var deploys []deployResult
	err = json.Unmarshal(response.Items, &deploys)
	if err != nil {
		return err
	}

	// TODO: Can we find the failed object?
	for _, item := range deploys {
		ads := item.ADS
		message := "Deployed %s in namespace %s to %s (%s)\n"
		if !item.Success {
			message = "Failed to deploy %s in namespace %s to %s (%s)\n"
		}
		fmt.Printf(message, ads.Name, ads.Namespace, ads.Cluster, item.DeployId)
	}

	return nil
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
