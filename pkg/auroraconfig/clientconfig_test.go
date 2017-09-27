package auroraconfig

import (
	"testing"

	"encoding/json"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

func TestResponse2ClientConfig(t *testing.T) {
	const expectedGitUrlPattern = "http://git-test"
	const expectedOpenShiftCluster = "utv"
	const expectedOpenShiftUrl = "https://utv-master"

	var err error

	var testClientConfig ClientConfig
	testClientConfig.GitUrlPattern = expectedGitUrlPattern
	testClientConfig.OpenShiftCluster = expectedOpenShiftCluster
	testClientConfig.OpenShiftUrl = expectedOpenShiftUrl

	var response serverapi.Response
	response.Success = true
	response.Message = "OK"
	response.Items = make([]json.RawMessage, 1)
	response.Items[0], err = json.Marshal(testClientConfig)
	if err != nil {
		t.Errorf("Error in Marshal testClientConfig: %v", err.Error())
	}
	response.Count = len(response.Items)

	clientConfig, err := response2ClientConfig(response)
	if err != nil {
		t.Errorf("Error in response2ClientConfig: %v", err.Error())
	}
	if clientConfig.GitUrlPattern != expectedGitUrlPattern {
		t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedGitUrlPattern, clientConfig.GitUrlPattern)
	}
	if clientConfig.OpenShiftCluster != expectedOpenShiftCluster {
		t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedOpenShiftCluster, clientConfig.OpenShiftCluster)
	}
	if clientConfig.OpenShiftUrl != expectedOpenShiftUrl {
		t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedOpenShiftUrl, clientConfig.OpenShiftUrl)
	}

}

func TestGetClientConfig(t *testing.T) {
	const expectedGitUrlPattern = "http://git-test"
	const expectedOpenShiftCluster = "utv"
	const expectedOpenShiftUrl = "https://utv-master"

	config := configuration.NewTestConfiguration()
	clientConfig, err := GetClientConfig(config)
	if err != nil {
		t.Errorf("Error in GetClientConfig: %v", err.Error())
	}
	if clientConfig == nil {
		t.Error("Could not get ClientConfig")
	} else {
		if clientConfig.GitUrlPattern != expectedGitUrlPattern {
			t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedGitUrlPattern, clientConfig.GitUrlPattern)
		}
		if clientConfig.OpenShiftCluster != expectedOpenShiftCluster {
			t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedOpenShiftCluster, clientConfig.OpenShiftCluster)
		}
		if clientConfig.OpenShiftUrl != expectedOpenShiftUrl {
			t.Errorf("Error in GitUrlPattern: Expected %v, got %v", expectedOpenShiftUrl, clientConfig.OpenShiftUrl)
		}
	}
}
