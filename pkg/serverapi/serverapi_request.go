package serverapi

import (
	"github.com/skatteetaten/ao/pkg/configuration"
)

type Request struct {
	ApiEndpoint 	string
	Method  		string
	Headers 		map[string]string
	Payload 		string
	Testing 		bool
}

func CallApiWithRequest(request *Request, config *configuration.ConfigurationClass) (result Response, err error) {

	return CallApiWithHeaders(request.Headers, request.Method, request.ApiEndpoint, request.Payload, config)
}