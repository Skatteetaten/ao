package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_deployToReachableClusters(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := ""

	deployClientMock := client.NewApplicationDeploymentClientMock()

	getClient := func(partition Partition) client.ApplicationDeploymentClient {
		return deployClientMock
	}

	partitions := []DeploySpecPartition{
		*newDeploySpecPartition(testSpecs[0:3], *newTestCluster("east", true), auroraConfig, overrideToken),
		*newDeploySpecPartition(testSpecs[3:7], *newTestCluster("west", true), auroraConfig, overrideToken),
		*newDeploySpecPartition(testSpecs[7:11], *newTestCluster("west", true), auroraConfig, overrideToken),
		*newDeploySpecPartition(testSpecs[11:13], *newTestCluster("north", true), auroraConfig, overrideToken),
	}

	deployClientMock.On("Deploy", mock.Anything).Times(4)

	_, err := deployToReachableClusters(getClient, partitions, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	deployClientMock.AssertExpectations(t)
}

func Test_deployToUnreachableClusters(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := ""

	deployClientMock := client.NewApplicationDeploymentClientMock()

	getClient := func(partition Partition) client.ApplicationDeploymentClient {
		return deployClientMock
	}

	partitions := []DeploySpecPartition{
		*newDeploySpecPartition(testSpecs[0:3], *newTestCluster("east", false), auroraConfig, overrideToken),
	}

	results, err := deployToReachableClusters(getClient, partitions, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, results, 1)
	assert.Len(t, results[0].Results, 3)
	assert.Equal(t, results[0].Results[0].Success, false)
	assert.Equal(t, results[0].Results[0].Reason, "Cluster is not reachable")

	assert.Equal(t, results[0].Results[1].Success, false)
	assert.Equal(t, results[0].Results[1].Reason, "Cluster is not reachable")

	assert.Equal(t, results[0].Results[2].Success, false)
	assert.Equal(t, results[0].Results[2].Reason, "Cluster is not reachable")
}
