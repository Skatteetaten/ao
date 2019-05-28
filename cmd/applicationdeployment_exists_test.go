package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
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
			[]client.DeploySpec{
				MockDeploySpec{"applicationDeploymentRef": "dev/crm", "cluster": "east", "envName": "dev", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "dev/erp", "cluster": "east", "envName": "dev", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "dev/sap", "cluster": "east", "envName": "dev", "affiliation": "sales"},
			},
			*newTestCluster("east", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]client.DeploySpec{
				MockDeploySpec{"applicationDeploymentRef": "test-qa/crm", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "test-qa/crmv2", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "test-qa/booking", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]client.DeploySpec{
				MockDeploySpec{"applicationDeploymentRef": "test-st/crm-1-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "test-st/crm-2-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "test-st/booking", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "test-st/erp", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
			},
			*newTestCluster("west", true),
			auroraConfigName,
			overrideToken),
		*newDeploySpecPartition(
			[]client.DeploySpec{
				MockDeploySpec{"applicationDeploymentRef": "prod/crm", "cluster": "north", "envName": "prod", "affiliation": "sales"},
				MockDeploySpec{"applicationDeploymentRef": "prod/booking", "cluster": "north", "envName": "prod", "affiliation": "sales"},
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
