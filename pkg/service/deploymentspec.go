package service

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
)

// GetFilteredDeploymentSpecs gets a filtered array of deployment specifications
func GetFilteredDeploymentSpecs(apiClient client.DeploySpecClient, applications []string, overrideCluster string) ([]deploymentspec.DeploymentSpec, error) {
	deploySpecs, err := apiClient.GetAuroraDeploySpec(applications, true)
	if err != nil {
		return nil, err
	}
	var filteredDeploymentSpecs []deploymentspec.DeploymentSpec
	if overrideCluster != "" {
		for _, spec := range deploySpecs {
			if spec.Cluster() == overrideCluster {
				filteredDeploymentSpecs = append(filteredDeploymentSpecs, spec)
			}
		}
	} else {
		filteredDeploymentSpecs = deploySpecs
	}

	return filteredDeploymentSpecs, nil
}
