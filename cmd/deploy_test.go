package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDeploySpec map[string]interface{}

type APIClientMock struct {
	mock.Mock
}

type AurorConfigClientMock struct {
	APIClientMock
	files []string
}

type DeploySpecClientMock struct {
	APIClientMock
	deploySpecs []client.DeploySpec
}

type DeployClientMock struct {
	APIClientMock
}

var applicationNames = [...]string{
	"dev/crm", "dev/erp", "dev/sap",
	"test-qa/crm", "test-qa/crmv2", "test-qa/booking", "test-qa/erp",
	"test-st/crm-1-GA", "test-st/crm-2-GA", "test-st/booking", "test-st/erp",
	"prod/crm", "prod/booking",
}

var fileNames = [...]string{
	"dev/crm.json", "dev/erp.json", "dev/sap.json", "dev/about.json",
	"test-qa/crm.json", "test-qa/crmv2.json", "test-qa/booking.json", "test-qa/erp.json", "test-qa/about.json",
	"test-st/crm-1-GA.json", "test-st/crm-2-GA.json", "test-st/booking.json", "test-st/erp.json", "test-st/about.json",
	"prod/crm.json", "prod/booking.json", "prod/about.json",
}

var testSpecs = [...]client.DeploySpec{
	&MockDeploySpec{"applicationId": "dev/crm", "cluster": "east", "envName": "dev"},
	&MockDeploySpec{"applicationId": "dev/erp", "cluster": "east", "envName": "dev"},
	&MockDeploySpec{"applicationId": "dev/sap", "cluster": "east", "envName": "dev"},
	&MockDeploySpec{"applicationId": "test-qa/crm", "cluster": "west", "envName": "test-qa"},
	&MockDeploySpec{"applicationId": "test-qa/crmv2", "cluster": "west", "envName": "test-qa"},
	&MockDeploySpec{"applicationId": "test-qa/booking", "cluster": "west", "envName": "test-qa"},
	&MockDeploySpec{"applicationId": "test-qa/erp", "cluster": "west", "envName": "test-qa"},
	&MockDeploySpec{"applicationId": "test-st/crm-1-GA", "cluster": "west", "envName": "test-st"},
	&MockDeploySpec{"applicationId": "test-st/crm-2-GA", "cluster": "west", "envName": "test-st"},
	&MockDeploySpec{"applicationId": "test-st/booking", "cluster": "west", "envName": "test-st"},
	&MockDeploySpec{"applicationId": "test-st/erp", "cluster": "west", "envName": "test-st"},
	&MockDeploySpec{"applicationId": "prod/crm", "cluster": "north", "envName": "prod"},
	&MockDeploySpec{"applicationId": "prod/booking", "cluster": "north", "envName": "prod"},
}

var testClusters = map[string]*config.Cluster{
	"east":  newTestCluster("east", true),
	"west":  newTestCluster("west", true),
	"north": newTestCluster("north", true),
}

func newAuroraConfigClientMock() *AurorConfigClientMock {
	return &AurorConfigClientMock{files: fileNames[:]}
}

func newDeploySpecClientMock() *DeploySpecClientMock {
	return &DeploySpecClientMock{deploySpecs: testSpecs[:]}
}

func newDeployClientMock() *DeployClientMock {
	return &DeployClientMock{}
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

func (api *APIClientMock) Do(method string, endpoint string, payload []byte) (*client.BooberResponse, error) {
	return nil, nil
}

func (api *APIClientMock) DoWithHeader(method string, endpoint string, header map[string]string, payload []byte) (*client.ResponseBundle, error) {
	return nil, nil
}

func (api *DeploySpecClientMock) GetAuroraDeploySpec(applications []string, defaults bool) ([]client.DeploySpec, error) {
	return api.deploySpecs, nil
}

func (api *DeploySpecClientMock) GetAuroraDeploySpecFormatted(environment, application string, defaults bool) (string, error) {
	return "", nil
}

func (api *AurorConfigClientMock) GetFileNames() (client.FileNames, error) {
	return api.files, nil
}

func (api *AurorConfigClientMock) GetAuroraConfig() (*client.AuroraConfig, error) {
	return nil, nil
}

func (api *AurorConfigClientMock) GetAuroraConfigNames() (*client.AuroraConfigNames, error) {
	return nil, nil
}

func (api *AurorConfigClientMock) PutAuroraConfig(endpoint string, ac *client.AuroraConfig) error {
	return nil
}

func (api *AurorConfigClientMock) ValidateAuroraConfig(ac *client.AuroraConfig, fullValidation bool) error {
	return nil
}

func (api *AurorConfigClientMock) PatchAuroraConfigFile(fileName string, operation client.JsonPatchOp) error {
	return nil
}

func (api *AurorConfigClientMock) GetAuroraConfigFile(fileName string) (*client.AuroraConfigFile, string, error) {
	return nil, "nil", nil
}

func (api *AurorConfigClientMock) PutAuroraConfigFile(file *client.AuroraConfigFile, eTag string) error {
	return nil
}

func (api *DeployClientMock) Deploy(deployPayload *client.DeployPayload) (*client.DeployResults, error) {
	api.Called()
	return &client.DeployResults{Message: "Successful", Success: true, Results: []client.DeployResult{}}, nil
}

func (api *DeployClientMock) GetApplyResult(deployId string) (string, error) {
	return "", nil
}

func (mds MockDeploySpec) Value(jsonPointer string) interface{} {
	key := strings.Replace(jsonPointer, "/", "", -1)
	return mds[key]
}

func Test_getApplications(t *testing.T) {
	search := "test/crm"

	apiClient := newAuroraConfigClientMock()

	actualApplications, err := getApplications(apiClient, search, "", []string{}, os.Stdout)
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

	apiClient := newAuroraConfigClientMock()

	actualApplications, err := getApplications(apiClient, search, "", exclusions, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actualApplications, 2)
	assert.Contains(t, actualApplications, "test-qa/crm")
	assert.Contains(t, actualApplications, "test-st/crm-2-GA")
}

func Test_getFilteredDeploymentSpecs(t *testing.T) {
	apiClient := newDeploySpecClientMock()

	filteredSpecs, err := getFilteredDeploymentSpecs(apiClient, applicationNames[:], "")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, filteredSpecs, 13)
}

func Test_getFilteredDeploymentSpecsWithOverrideCluster(t *testing.T) {
	apiClient := newDeploySpecClientMock()

	filteredSpecs, err := getFilteredDeploymentSpecs(apiClient, applicationNames[:], "east")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, filteredSpecs, 3)
}

func Test_createDeploymentUnits(t *testing.T) {

	affiliation := "sales"
	overrideToken := ""

	units := createDeploymentUnits(affiliation, overrideToken, testClusters, testSpecs[:])

	assert.Len(t, units, 4)
	assert.Contains(t, units, *newDeploymentUnitID("east", "dev"))
	assert.Contains(t, units, *newDeploymentUnitID("west", "test-st"))
	assert.Contains(t, units, *newDeploymentUnitID("west", "test-qa"))
	assert.Contains(t, units, *newDeploymentUnitID("north", "prod"))

	sampleUnit := units[*newDeploymentUnitID("west", "test-st")]

	assert.Equal(t, affiliation, sampleUnit.affiliation)
	assert.Equal(t, "west", sampleUnit.cluster.Name)
	assert.Equal(t, overrideToken, sampleUnit.overrideToken)
	assert.Len(t, sampleUnit.applicationList, 4)
	assert.Contains(t, sampleUnit.applicationList, "test-st/crm-1-GA")
	assert.Contains(t, sampleUnit.applicationList, "test-st/crm-2-GA")
	assert.Contains(t, sampleUnit.applicationList, "test-st/booking")
	assert.Contains(t, sampleUnit.applicationList, "test-st/erp")
}

func Test_createDeploymentUnitsWithOverrideToken(t *testing.T) {
	affiliation := "east"
	overrideToken := "footoken"

	units := createDeploymentUnits(affiliation, overrideToken, testClusters, testSpecs[:])

	sampleUnit := units[*newDeploymentUnitID("west", "test-st")]

	assert.Equal(t, overrideToken, sampleUnit.overrideToken)
}

func Test_deployToReachableClusters(t *testing.T) {
	affiliation := "sales"
	overrideToken := ""

	deployClientMock := newDeployClientMock()

	getClient := func(unit *deploymentUnit) client.DeployClient {
		return deployClientMock
	}

	deploymentUnits := map[deploymentUnitID]*deploymentUnit{
		*newDeploymentUnitID("east", "dev"): newDeploymentUnit(newDeploymentUnitID("east", "dev"),
			[]string{"dev/crm", "dev/erp", "dev/sap"}, newTestCluster("east", true), affiliation, overrideToken),
		*newDeploymentUnitID("west", "test-qa"): newDeploymentUnit(newDeploymentUnitID("west", "test"),
			[]string{"test-qa/crm", "test-qa/crmv2", "test-qa/booking", "test-qa/erp"}, newTestCluster("west", true), affiliation, overrideToken),
		*newDeploymentUnitID("west", "test-st"): newDeploymentUnit(newDeploymentUnitID("west", "test"),
			[]string{"test-st/crm-1-GA", "test-st/crm-2-GA", "test-st/booking", "test-st/erp"}, newTestCluster("west", true), affiliation, overrideToken),
		*newDeploymentUnitID("north", "prod"): newDeploymentUnit(newDeploymentUnitID("north", "prod"),
			[]string{"prod/crm", "prod/booking"}, newTestCluster("north", true), affiliation, overrideToken),
	}

	deployClientMock.On("Deploy", mock.Anything).Times(4)

	_, err := deployToReachableClusters(getClient, deploymentUnits, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	deployClientMock.AssertExpectations(t)
}
