package cmd

import (
	"testing"

	"ao/pkg/client"
	"ao/pkg/deploymentspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_checkExistence(t *testing.T) {
	auroraConfigName := "jupiter"
	overrideToken := ""

	applicationDeploymentClientMock := client.NewApplicationDeploymentClientMock()

	getClient := func(partition Partition) client.ApplicationDeploymentClient {
		return applicationDeploymentClientMock
	}

	partitions := []DeploySpecPartition{
		*newDeploySpecPartition(
			[]deploymentspec.DeploymentSpec{
				deploymentspec.NewDeploymentSpec("crm", "dev", "east", "1"),
				deploymentspec.NewDeploymentSpec("erp", "dev", "east", "1"),
				deploymentspec.NewDeploymentSpec("booking", "dev", "east", "1"),
			},
			*newTestCluster("east", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]deploymentspec.DeploymentSpec{
				deploymentspec.NewDeploymentSpec("crm", "test-qa", "west", "1"),
				deploymentspec.NewDeploymentSpec("crmv2", "test-qa", "west", "1"),
				deploymentspec.NewDeploymentSpec("booking", "test-qa", "west", "1"),
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]deploymentspec.DeploymentSpec{
				deploymentspec.NewDeploymentSpec("crm-1-GA", "test-st", "west", "1"),
				deploymentspec.NewDeploymentSpec("crm-2-GA", "test-st", "west", "1"),
				deploymentspec.NewDeploymentSpec("booking", "test-st", "west", "1"),
				deploymentspec.NewDeploymentSpec("bookingv2", "test-st", "west", "1"),
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]deploymentspec.DeploymentSpec{
				deploymentspec.NewDeploymentSpec("crm", "prod", "north", "1"),
				deploymentspec.NewDeploymentSpec("booking", "prod", "north", "1"),
			},
			*newTestCluster("north", true),
			auroraConfigName,
			overrideToken),
	}

	applicationDeploymentClientMock.On("Exists", mock.Anything).Times(4)

	results, err := checkExistence(getClient, partitions)
	if err != nil {
		t.Fatal(err)
	}

	applicationDeploymentClientMock.AssertExpectations(t)
	assert.Len(t, results, 4)
}
