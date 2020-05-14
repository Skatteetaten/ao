package cmd

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/stretchr/testify/assert"
)

var testClusters = map[string]*config.Cluster{
	"east":  newTestCluster("east", true),
	"west":  newTestCluster("west", true),
	"north": newTestCluster("north", true),
}

var testSpecs = [...]deploymentspec.DeploymentSpec{
	deploymentspec.NewDeploymentSpec("crm", "dev", "east", "1"),
	deploymentspec.NewDeploymentSpec("erp", "dev", "east", "1"),
	deploymentspec.NewDeploymentSpec("sap", "dev", "east", "1"),
	deploymentspec.NewDeploymentSpec("crm", "test-qa", "west", "1"),
	deploymentspec.NewDeploymentSpec("crmv2", "test-qa", "west", "1"),
	deploymentspec.NewDeploymentSpec("booking", "test-qa", "west", "1"),
	deploymentspec.NewDeploymentSpec("erp", "test-qa", "west", "1"),
	deploymentspec.NewDeploymentSpec("crm-1-GA", "test-st", "west", "1"),
	deploymentspec.NewDeploymentSpec("crm-2-GA", "test-st", "west", "1"),
	deploymentspec.NewDeploymentSpec("booking", "test-st", "west", "1"),
	deploymentspec.NewDeploymentSpec("erp", "test-st", "west", "1"),
	deploymentspec.NewDeploymentSpec("crm", "prod", "north", "1"),
	deploymentspec.NewDeploymentSpec("booking", "prod", "north", "1"),
}

func newTestCluster(name string, reachable bool) *config.Cluster {
	return &config.Cluster{
		Name:      name,
		URL:       name + ".url",
		Token:     name + ".token",
		Reachable: reachable,
		BooberURL: name + "boober.url",
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
