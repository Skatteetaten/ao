package service

import (
	"strings"
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/stretchr/testify/assert"
)

type MockDeploySpec map[string]interface{}

var applicationNames = [...]string{
	"dev/crm", "dev/erp", "dev/sap",
	"test-qa/crm", "test-qa/crmv2", "test-qa/booking", "test-qa/erp",
	"test-st/crm-1-GA", "test-st/crm-2-GA", "test-st/booking", "test-st/erp",
	"prod/crm", "prod/booking",
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

func (mds MockDeploySpec) Value(jsonPointer string) interface{} {
	key := strings.Replace(jsonPointer, "/", "", -1)
	return mds[key]
}

func Test_getFilteredDeploymentSpecs(t *testing.T) {
	apiClient := client.NewDeploySpecClientMock(testSpecs[:])

	filteredSpecs, err := GetFilteredDeploymentSpecs(apiClient, applicationNames[:], "")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, filteredSpecs, 13)
}

func Test_getFilteredDeploymentSpecsWithOverrideCluster(t *testing.T) {
	apiClient := client.NewDeploySpecClientMock(testSpecs[:])

	filteredSpecs, err := GetFilteredDeploymentSpecs(apiClient, applicationNames[:], "east")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, filteredSpecs, 3)
}
