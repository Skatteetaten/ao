package client

import (
	"errors"
	"strings"
)

const getApplicationDeploymentNamespace = `
        query applicationDeployment($id: String!) {
                applicationDeployment(id: $id) {
                        namespace {
                                name
                        }
                }
        }
`

type ApplicationDeploymentNamespaceResponse struct {
	ApplicationDeployment struct {
		Namespace struct {
			Name string `json:"name"`
		} `json:"namespace"`
	} `json:"applicationDeployment"`
}

func (api *APIClient) GetNamespace(applicationDeploymentID string) (string, error) {
	vars := map[string]interface{}{
		"id": applicationDeploymentID,
	}

	var adNamespaceResponse ApplicationDeploymentNamespaceResponse

	if err := api.RunGraphQl(getApplicationDeploymentNamespace, vars, &adNamespaceResponse); err != nil {
		if strings.Contains(err.Error(), "The requested resource was not found") {
			return "", errors.New("The application is not deployed. Please deploy it with ao deploy.")
		}
		return "", err
	}

	return adNamespaceResponse.ApplicationDeployment.Namespace.Name, nil
}
