package serverapi

import (
	"encoding/json"

	"errors"
	"io/ioutil"
	"strings"

	"github.com/skatteetaten/ao/pkg/configuration"
)

type Request struct {
	ApiEndpoint string
	Method      string
	Headers     map[string]string
	Payload     string
	Testing     bool
}

func CallApiWithRequest(request *Request, config *configuration.ConfigurationClass) (result Response, err error) {
	if config.Testing {
		return generateTestResponse(request.Method, request.ApiEndpoint, request.Payload)
	} else {
		return CallApiWithHeaders(request.Headers, request.Method, request.ApiEndpoint, request.Payload, config)
	}
}

func generateTestResponse(method string, apiEndpoint string, Payload string) (result Response, err error) {
	testFileName, err := getTestFileName(apiEndpoint)
	if err != nil {
		return result, err
	}

	content, err := ioutil.ReadFile(testFileName)
	if err != nil {
		return result, err
	}

	result.Items = make([]json.RawMessage, 1)
	result.Items[0] = json.RawMessage(content)
	result.Success = true
	result.Count = 1
	result.Message = "Success"

	return result, nil
}

func getTestFileName(ApiEndpoint string) (testFileName string, err error) {
	if strings.Contains(ApiEndpoint, "auroraconfig") {
		testFileName = "../serverapi/Testfiles/auroraconfig.json"
	} else if strings.Contains(ApiEndpoint, "vault") {
		testFileName = "../serverapi/Testfiles/vault.json"
	} else {
		err = errors.New("Illegal API endpoint for testing: " + ApiEndpoint)
		return "", err
	}
	return testFileName, nil
}
