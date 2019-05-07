package service

import (
	"strings"
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
)

type MockDeploySpec map[string]interface{}

var applicationNames = [...]string{
	"dev/crm", "dev/erp", "dev/sap",
	"test-qa/crm", "test-qa/crmv2", "test-qa/booking", "test-qa/erp",
	"test-st/crm-1-GA", "test-st/crm-2-GA", "test-st/booking", "test-st/erp",
	"prod/crm", "prod/booking",
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
