package client

// DeploySpecClientMock is a base mock type
type DeploySpecClientMock struct {
	APIClientMock
	deploySpecs []DeploySpec
}

// NewDeploySpecClientMock creates a new DeploySpecClientMock
func NewDeploySpecClientMock(deploySpecs []DeploySpec) *DeploySpecClientMock {
	return &DeploySpecClientMock{deploySpecs: deploySpecs}
}

// GetAuroraDeploySpec default mock implementation
func (api *DeploySpecClientMock) GetAuroraDeploySpec(applications []string, defaults bool) ([]DeploySpec, error) {
	return api.deploySpecs, nil
}

// GetAuroraDeploySpecFormatted default mock implementation
func (api *DeploySpecClientMock) GetAuroraDeploySpecFormatted(environment, application string, defaults bool) (string, error) {
	return "", nil
}
