package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_deleteFromReachableClusters(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := ""

	applicationDeploymentClientMock := newApplicationDeploymentClientMock()

	getClient := func(partition *requestPartition) client.ApplicationDeploymentClient {
		return applicationDeploymentClientMock
	}

	partitions := map[requestPartitionID]*requestPartition{
		*newRequestPartitionID("east", "dev"): newRequestPartition(newRequestPartitionID("east", "dev"),
			testSpecs[0:3], newTestCluster("east", true), auroraConfig, overrideToken),
		*newRequestPartitionID("west", "test-qa"): newRequestPartition(newRequestPartitionID("west", "test"),
			testSpecs[3:7], newTestCluster("west", true), auroraConfig, overrideToken),
		*newRequestPartitionID("west", "test-st"): newRequestPartition(newRequestPartitionID("west", "test"),
			testSpecs[7:11], newTestCluster("west", true), auroraConfig, overrideToken),
		*newRequestPartitionID("north", "prod"): newRequestPartition(newRequestPartitionID("north", "prod"),
			testSpecs[11:13], newTestCluster("north", true), auroraConfig, overrideToken),
	}

	applicationDeploymentClientMock.On("Delete", mock.Anything).Times(4)

	results, err := deleteFromReachableClusters(getClient, partitions)
	if err != nil {
		t.Fatal(err)
	}

	applicationDeploymentClientMock.AssertExpectations(t)
	assert.Len(t, results, 4)
}
