package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_createDeploymentPartitions(t *testing.T) {

	auroraConfig := "jupiter"
	overrideToken := ""

	deploymentInfos := [...]DeploymentInfo{
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

	clusters := map[string]*config.Cluster{
		"east":  newTestCluster("east", true),
		"west":  newTestCluster("west", true),
		"north": newTestCluster("north", true),
	}

	partitions, err := createDeploymentPartitions(auroraConfig, overrideToken, clusters, deploymentInfos[:])

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, partitions, 4)
}

func Test_deleteFromReachableClusters(t *testing.T) {
	auroraConfigName := "jupiter"
	overrideToken := ""

	applicationDeploymentClientMock := client.NewApplicationDeploymentClientMock()

	getClient := func(partition Partition) client.ApplicationDeploymentClient {
		return applicationDeploymentClientMock
	}

	partitions := []DeploymentPartition{
		*newDeploymentPartition(
			[]DeploymentInfo{
				*newDeploymentInfo("sales-dev", "crm", "east"),
				*newDeploymentInfo("sales-dev", "erp", "east"),
				*newDeploymentInfo("sales-dev", "booking", "east"),
			},
			*newTestCluster("east", true),
			auroraConfigName,
			overrideToken),
		*newDeploymentPartition(
			[]DeploymentInfo{
				*newDeploymentInfo("sales-qa", "crm", "west"),
				*newDeploymentInfo("sales-qa", "erp", "west"),
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploymentPartition(
			[]DeploymentInfo{
				*newDeploymentInfo("finance-dev", "crm", "west"),
				*newDeploymentInfo("finance-dev", "crm-v2", "west"),
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploymentPartition(
			[]DeploymentInfo{
				*newDeploymentInfo("finance-qa", "erp", "west"),
				*newDeploymentInfo("finance-qa", "booking", "west"),
			},
			*newTestCluster("north", true),
			auroraConfigName,
			overrideToken),
	}

	applicationDeploymentClientMock.On("Delete", mock.Anything).Times(4)

	results, err := deleteFromReachableClusters(getClient, partitions)
	if err != nil {
		t.Fatal(err)
	}

	applicationDeploymentClientMock.AssertExpectations(t)
	assert.Len(t, results, 4)
}
