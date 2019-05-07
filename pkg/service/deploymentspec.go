package service

import "github.com/skatteetaten/ao/pkg/client"

func GetFilteredDeploymentSpecs(apiClient client.DeploySpecClient, applications []string, overrideCluster string) ([]client.DeploySpec, error) {
	deploySpecs, err := apiClient.GetAuroraDeploySpec(applications, true)
	if err != nil {
		return nil, err
	}
	var filteredDeploymentSpecs []client.DeploySpec
	if overrideCluster != "" {
		for _, spec := range deploySpecs {
			if spec.Value("/cluster").(string) == overrideCluster {
				filteredDeploymentSpecs = append(filteredDeploymentSpecs, spec)
			}
		}
	} else {
		filteredDeploymentSpecs = deploySpecs
	}

	return filteredDeploymentSpecs, nil
}
