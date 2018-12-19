package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
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
	&MockDeploySpec{"applicationId": "dev/crm", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "dev/erp", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "dev/sap", "cluster": "east", "envName": "dev", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-qa/crm", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-qa/crmv2", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-qa/booking", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-qa/erp", "cluster": "west", "envName": "test-qa", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-st/crm-1-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-st/crm-2-GA", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-st/booking", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "test-st/erp", "cluster": "west", "envName": "test-st", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "prod/crm", "cluster": "north", "envName": "prod", "affiliation": "sales"},
	&MockDeploySpec{"applicationId": "prod/booking", "cluster": "north", "envName": "prod", "affiliation": "sales"},
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
	return nil, errors.New("Not implemented")
}

func (api *AurorConfigClientMock) GetAuroraConfigNames() (*client.AuroraConfigNames, error) {
	return nil, errors.New("Not implemented")
}

func (api *AurorConfigClientMock) PutAuroraConfig(endpoint string, ac *client.AuroraConfig) error {
	return errors.New("Not implemented")
}

func (api *AurorConfigClientMock) ValidateAuroraConfig(ac *client.AuroraConfig, fullValidation bool) error {
	return errors.New("Not implemented")
}

func (api *AurorConfigClientMock) PatchAuroraConfigFile(fileName string, operation client.JsonPatchOp) error {
	return errors.New("Not implemented")
}

func (api *AurorConfigClientMock) GetAuroraConfigFile(fileName string) (*client.AuroraConfigFile, string, error) {
	return nil, "", errors.New("Not implemented")
}

func (api *AurorConfigClientMock) PutAuroraConfigFile(file *client.AuroraConfigFile, eTag string) error {
	return errors.New("Not implemented")
}

func (api *DeployClientMock) Deploy(deployPayload *client.DeployPayload) (*client.DeployResults, error) {
	api.Called()
	return &client.DeployResults{Message: "Successful", Success: true, Results: []client.DeployResult{}}, nil
}

func (api *DeployClientMock) GetApplyResult(deployId string) (string, error) {
	return "", errors.New("Not implemented")
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

	auroraConfig := "jupiter"
	overrideToken := ""

	units := createDeploymentUnits(auroraConfig, overrideToken, testClusters, testSpecs[:])

	assert.Len(t, units, 4)
	assert.Contains(t, units, *newDeploymentUnitID("east", "dev"))
	assert.Contains(t, units, *newDeploymentUnitID("west", "test-st"))
	assert.Contains(t, units, *newDeploymentUnitID("west", "test-qa"))
	assert.Contains(t, units, *newDeploymentUnitID("north", "prod"))

	sampleUnit := units[*newDeploymentUnitID("west", "test-st")]

	assert.Equal(t, auroraConfig, sampleUnit.auroraConfig)
	assert.Equal(t, "west", sampleUnit.cluster.Name)
	assert.Equal(t, overrideToken, sampleUnit.overrideToken)
	assert.Len(t, sampleUnit.deploySpecList, 4)
	assert.IsType(t, sampleUnit.deploySpecList[0], &MockDeploySpec{})
	assert.IsType(t, sampleUnit.deploySpecList[1], &MockDeploySpec{})
	assert.IsType(t, sampleUnit.deploySpecList[2], &MockDeploySpec{})
	assert.IsType(t, sampleUnit.deploySpecList[3], &MockDeploySpec{})
}

func Test_createDeploymentUnitsWithOverrideToken(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := "footoken"

	units := createDeploymentUnits(auroraConfig, overrideToken, testClusters, testSpecs[:])

	sampleUnit := units[*newDeploymentUnitID("west", "test-st")]

	assert.Equal(t, overrideToken, sampleUnit.overrideToken)
}

func Test_deployToReachableClusters(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := ""

	deployClientMock := newDeployClientMock()

	getClient := func(unit *deploymentUnit) client.DeployClient {
		return deployClientMock
	}

	deploymentUnits := map[deploymentUnitID]*deploymentUnit{
		*newDeploymentUnitID("east", "dev"): newDeploymentUnit(newDeploymentUnitID("east", "dev"),
			testSpecs[0:3], newTestCluster("east", true), auroraConfig, overrideToken),
		*newDeploymentUnitID("west", "test-qa"): newDeploymentUnit(newDeploymentUnitID("west", "test"),
			testSpecs[3:7], newTestCluster("west", true), auroraConfig, overrideToken),
		*newDeploymentUnitID("west", "test-st"): newDeploymentUnit(newDeploymentUnitID("west", "test"),
			testSpecs[7:11], newTestCluster("west", true), auroraConfig, overrideToken),
		*newDeploymentUnitID("north", "prod"): newDeploymentUnit(newDeploymentUnitID("north", "prod"),
			testSpecs[11:13], newTestCluster("north", true), auroraConfig, overrideToken),
	}

	deployClientMock.On("Deploy", mock.Anything).Times(4)

	_, err := deployToReachableClusters(getClient, deploymentUnits, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	deployClientMock.AssertExpectations(t)
}

func Test_deployToUnreachableClusters(t *testing.T) {
	auroraConfig := "jupiter"
	overrideToken := ""

	deployClientMock := newDeployClientMock()

	getClient := func(unit *deploymentUnit) client.DeployClient {
		return deployClientMock
	}

	deploymentUnits := map[deploymentUnitID]*deploymentUnit{
		*newDeploymentUnitID("east", "dev"): newDeploymentUnit(newDeploymentUnitID("east", "dev"),
			testSpecs[0:3], newTestCluster("east", false), auroraConfig, overrideToken),
	}

	results, err := deployToReachableClusters(getClient, deploymentUnits, map[string]string{})
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
