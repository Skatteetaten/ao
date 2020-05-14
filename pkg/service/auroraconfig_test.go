package service

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

var fileNames = [...]string{
	"dev/crm.json", "dev/erp.json", "dev/sap.json", "dev/about.json",
	"test-qa/crm.json", "test-qa/crmv2.json", "test-qa/booking.json", "test-qa/erp.json", "test-qa/about.json",
	"test-st/crm-1-GA.json", "test-st/crm-2-GA.json", "test-st/booking.json", "test-st/erp.json", "test-st/about.json",
	"prod/crm.json", "prod/booking.json", "prod/about.json",
}

func Test_getApplications(t *testing.T) {
	search := "test/crm"

	apiClient := client.NewAuroraConfigClientMock(fileNames[:])

	actualApplications, err := GetApplications(apiClient, search, []string{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actualApplications, 4)
	assert.Contains(t, actualApplications, "test-qa/crm")
	assert.Contains(t, actualApplications, "test-qa/crmv2")
	assert.Contains(t, actualApplications, "test-st/crm-1-GA")
	assert.Contains(t, actualApplications, "test-st/crm-2-GA")
}

func Test_getApplicationsWithExclusions(t *testing.T) {
	search := "test/crm"
	exclusions := []string{"test-qa/crmv2", "test-st/crm-1-GA"}

	apiClient := client.NewAuroraConfigClientMock(fileNames[:])

	actualApplications, err := GetApplications(apiClient, search, exclusions)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actualApplications, 2)
	assert.Contains(t, actualApplications, "test-qa/crm")
	assert.Contains(t, actualApplications, "test-st/crm-2-GA")
}
