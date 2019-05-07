package client

import (
	"github.com/pkg/errors"
)

// ApplicationDeploymentClientMock is a base mock type
type ApplicationDeploymentClientMock struct {
	APIClientMock
}

// NewApplicationDeploymentClientMock creates a new ApplicationDeploymentClientMock
func NewApplicationDeploymentClientMock() *ApplicationDeploymentClientMock {
	return &ApplicationDeploymentClientMock{}
}

// Deploy default mock implementation
func (api *ApplicationDeploymentClientMock) Deploy(deployPayload *DeployPayload) (*DeployResults, error) {
	api.Called()
	return &DeployResults{Message: "Successful", Success: true, Results: []DeployResult{}}, nil
}

// Delete default mock implementation
func (api *ApplicationDeploymentClientMock) Delete(deletePayload *DeletePayload) (*DeleteResults, error) {
	api.Called()
	return &DeleteResults{Message: "Successful", Success: true, Results: []DeleteResult{}}, nil
}

// Exists default mock implementation
func (api *ApplicationDeploymentClientMock) Exists(existsPayload *ExistsPayload) (*ExistsResults, error) {
	api.Called()
	return &ExistsResults{Message: "Successful", Success: true, Results: []ExistsResult{}}, nil
}

// GetApplyResult default mock implementation
func (api *ApplicationDeploymentClientMock) GetApplyResult(deployID string) (string, error) {
	return "", errors.New("Not implemented")
}
