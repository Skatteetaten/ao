package client

import "ao/pkg/deploymentspec"

// DeploySpecClientMock is a base mock type
type DeploySpecClientMock struct {
	APIClientMock
	deploySpecs []deploymentspec.DeploymentSpec
}

// NewDeploySpecClientMock creates a new DeploySpecClientMock
func NewDeploySpecClientMock(deploySpecs []deploymentspec.DeploymentSpec) *DeploySpecClientMock {
	return &DeploySpecClientMock{deploySpecs: deploySpecs}
}

// GetAuroraDeploySpec default mock implementation
func (api *DeploySpecClientMock) GetAuroraDeploySpec(applications []string, defaults bool) ([]deploymentspec.DeploymentSpec, error) {
	return api.deploySpecs, nil
}

// GetAuroraDeploySpecFormatted default mock implementation
func (api *DeploySpecClientMock) GetAuroraDeploySpecFormatted(environment, application string, defaults bool) (string, error) {
	return "", nil
}
