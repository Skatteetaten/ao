package serverapi

import (
	"encoding/json"

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
		return generateTestRequest(request.Method, request.ApiEndpoint, request.Payload)
	} else {
		return CallApiWithHeaders(request.Headers, request.Method, request.ApiEndpoint, request.Payload, config)
	}
}

func generateTestRequest(method string, ApiEndpoint string, Payload string) (result Response, err error) {
	result.Items = make([]json.RawMessage, 0)
	result.Success = true
	result.Count = 0
	result.Message = "Testing"
	return
}
