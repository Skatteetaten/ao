package boober

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/sirupsen/logrus"
)

type deployResponse struct {
	Response
	Items []struct {
		DeployId string `json:"deployId"`
		ADS struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Cluster   string `json:"cluster"`
		} `json:"auroraDeploymentSpec"`
		Success bool `json:"success"`
	} `json:"items"`
}

func (api *BooberClient) Deploy(applications []string, overrides map[string]json.RawMessage) (*Validation) {

	applicationIds := createApplicationIds(applications)

	applyPayload := struct {
		ApplicationIds []ApplicationId            `json:"applicationIds"`
		Overrides      map[string]json.RawMessage `json:"overrides"`
		Deploy         bool                       `json:"deploy"`
	}{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
		Deploy:         true,
	}

	payload, err := json.Marshal(applyPayload)
	if err != nil {
		fmt.Println("Failed to marshal DeployPayload")
		return nil
	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)

	logrus.Info("Deploying to ", api.Host)
	var response deployResponse
	validation, err := api.Call(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &response)
		return response, jErr
	})
	if err != nil {
		fmt.Println(err)
		return validation
	}

	for _, item := range response.Items {
		if !item.Success {
			fmt.Printf("Failed to deploy %s/%s to %s (%s)\n", item.ADS.Namespace, item.ADS.Name, item.ADS.Cluster, item.DeployId)
		} else {
			fmt.Printf("Deployed %s in namespace %s to %s (%s)\n", item.ADS.Name, item.ADS.Namespace, item.ADS.Cluster, item.DeployId)
		}
	}

	return validation
}

func createApplicationIds(apps []string) []ApplicationId {
	applicationIds := []ApplicationId{}
	for _, app := range apps {
		envApp := strings.Split(app, "/")

		if len(envApp) != 2 {
			continue
		}
		applicationIds = append(applicationIds, ApplicationId{
			Environment: envApp[0],
			Application: envApp[1],
		})
	}
	return applicationIds
}
