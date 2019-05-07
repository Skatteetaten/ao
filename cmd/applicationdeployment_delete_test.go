package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testDeploymentInfos = [...]DeploymentInfo{
	*newDeploymentInfo("sales-dev", "crm", "east"),
	*newDeploymentInfo("sales-dev", "erp", "east"),
	*newDeploymentInfo("sales-dev", "booking", "east"),
	*newDeploymentInfo("sales-qa", "crm", "west"),
	*newDeploymentInfo("sales-qa", "erp", "west"),
	*newDeploymentInfo("finance-dev", "crm", "west"),
	*newDeploymentInfo("finance-dev", "crm-v2", "west"),
	*newDeploymentInfo("finance-qa", "erp", "west"),
	*newDeploymentInfo("finance-qa", "booking", "west"),
}

func Test_deleteFromReachableClusters(t *testing.T) {
	overrideToken := ""

	applicationDeploymentClientMock := client.NewApplicationDeploymentClientMock()

	getClient := func(partition Partition) client.ApplicationDeploymentClient {
		return applicationDeploymentClientMock
	}

	partitions := []DeploymentPartition{
		*newDeploymentPartition(testDeploymentInfos[0:3], *newTestCluster("east", true), "jupiter", overrideToken),
		*newDeploymentPartition(testDeploymentInfos[3:5], *newTestCluster("west", true), "jupiter", overrideToken),
		*newDeploymentPartition(testDeploymentInfos[5:7], *newTestCluster("west", true), "jupiter", overrideToken),
		*newDeploymentPartition(testDeploymentInfos[7:9], *newTestCluster("north", true), "jupiter", overrideToken),
	}

	applicationDeploymentClientMock.On("Delete", mock.Anything).Times(4)

	results, err := deleteFromReachableClusters(getClient, partitions)
	if err != nil {
		t.Fatal(err)
	}

	applicationDeploymentClientMock.AssertExpectations(t)
	assert.Len(t, results, 4)
}
