package cmd

import (
	"strings"
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/stretchr/testify/assert"
)

type MockDeploySpec map[string]interface{}

var testClusters = map[string]*config.Cluster{
	"east":  newTestCluster("east", true),
	"west":  newTestCluster("west", true),
	"north": newTestCluster("north", true),
}

var testSpecs = [...]client.DeploySpec{
	&MockDeploySpec{"applicationDeploymentRef": "dev/crm", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "dev/erp", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "dev/sap", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-qa/crm", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-qa/crmv2", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-qa/booking", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-qa/erp", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-st/crm-1-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-st/crm-2-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-st/booking", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "test-st/erp", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "prod/crm", "cluster": "north", "envName": "prod", "affiliation": "sales"},
	&MockDeploySpec{"applicationDeploymentRef": "prod/booking", "cluster": "north", "envName": "prod", "affiliation": "sales"},
}

func (mds MockDeploySpec) Value(jsonPointer string) interface{} {
	key := strings.Replace(jsonPointer, "/", "", -1)
	return mds[key]
}

func newTestCluster(name string, reachable bool) *config.Cluster {
	return &config.Cluster{
		Name:      name,
		Url:       name + ".url",
		Token:     name + ".token",
		Reachable: reachable,
		BooberUrl: name + "boober.url",
	}
}

func Test_createDeploySpecPartitions(t *testing.T) {

	auroraConfig := "jupiter"
	overrideToken := ""

	partitions, err := createDeploySpecPartitions(auroraConfig, overrideToken, testClusters, testSpecs[:])

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, partitions, 4)
}

func Test_createRequestPartitionsWithOverrideToken(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := "footoken"

	partitions, err := createDeploySpecPartitions(auroraConfig, overrideToken, testClusters, testSpecs[:])

	if err != nil {
		t.Fatal(err)
	}

	samplePartition := partitions[0]

	assert.Equal(t, overrideToken, samplePartition.OverrideToken)
}
