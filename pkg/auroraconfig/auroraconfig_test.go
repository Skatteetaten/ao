package auroraconfig

import (
	"encoding/json"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestGetAuroraConfigRequest(t *testing.T) {
	var config *configuration.ConfigurationClass = new(configuration.ConfigurationClass)
	config.OpenshiftConfig = new(openshift.OpenshiftConfig)
	config.OpenshiftConfig.Affiliation = "foobar"

	request := GetAuroraConfigRequest(config)
	const expectedApiEndpoint = "/affiliation/foobar/auroraconfig"
	if request.ApiEndpoint != expectedApiEndpoint {
		t.Errorf("Unexpected API Endpoint, expected %v, got %v", expectedApiEndpoint, request.ApiEndpoint)
	}

	if request.Method != http.MethodGet {
		t.Errorf("Unexpected method, expected GET, got %v", request.Method)
	}
}

func TestResponse2AuroraConfig(t *testing.T) {
	var response serverapi.Response
	response.Count = 1
	response.Items = make([]json.RawMessage, 1)

	jsonContent, err := ioutil.ReadFile("Testfiles/auroraconfig1file.json")
	if err != nil {
		t.Errorf("Error in reading testfile: %v", err.Error())
	}

	response.Items[0] = json.RawMessage(jsonContent)
	response.Success = true
	response.Message = "OK"

	auroraConfig, err := Response2AuroraConfig(response)
	if err != nil {
		t.Errorf("Error: %v", err.Error())
	}
	if len(auroraConfig.Files) != 1 {
		t.Errorf("Expected 1 file, got %v", len(auroraConfig.Files))
	}

	response.Success = false
	message := "Something went wrong"
	stripMessage := jsonutil.StripSpaces("Something went wrong")
	response.Message = message
	auroraConfig, err = Response2AuroraConfig(response)
	if err == nil {
		t.Errorf("Error response did not trigger error")
	} else {
		stripError := jsonutil.StripSpaces(err.Error())
		if !strings.Contains(stripError, stripMessage) {
			t.Errorf("Unexpected error message, expected \"%v\", got \"%v\"", stripMessage, stripError)
		}
	}

}
