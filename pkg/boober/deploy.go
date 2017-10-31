package boober

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DeployResponse struct {
	Response
	Items []struct {
		DeployId string `json:"deployId"`
		ADS struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"auroraDeploymentSpec"`
		Success bool `json:"success"`
	} `json:"items"`
}

func (api *Api) Deploy(applicationIds []ApplicationId, overrides map[string]json.RawMessage) error {

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
		return err
	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)

	var response DeployResponse
	api.WithRequest(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &response)
		return response, jErr
	})

	if err != nil {
		return err
	}

	for _, item := range response.Items {
		if !item.Success {
			fmt.Printf("Failed to deploy: %s/%s", item.ADS.Namespace, item.ADS.Name)
		}
		fmt.Printf("Deployed: %s/%s\n", item.ADS.Namespace, item.ADS.Name)
	}

	return nil
}
