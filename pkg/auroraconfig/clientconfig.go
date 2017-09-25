package auroraconfig

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

type ClientConfig struct {
	GitUrlPattern    string `json:"gitUrlPattern"`
	OpenShiftCluster string `json:"openshiftCluster"`
	OpenShiftUrl     string `json:"openshiftUrl"`
}

type ClientConfigResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Items   []ClientConfig `json:"items"`
	Count   int            `json:"count"`
}

func GetClientConfigRequest(configuration *configuration.ConfigurationClass) (request *serverapi.Request) {
	request = new(serverapi.Request)
	request.Method = http.MethodGet
	request.ApiEndpoint = "/clientconfig/"
	return request
}

func response2ClientConfig(response serverapi.Response) (clientConfig *ClientConfig, err error) {
	clientConfig = new(ClientConfig)
	if len(response.Items) != 1 {
		err = errors.New("Internal error: None or Multiple Client Config received")
		return clientConfig, err
	}
	err = json.Unmarshal(response.Items[0], &clientConfig)
	return clientConfig, err
}

func GetClientConfig(config *configuration.ConfigurationClass) (clientConfig *ClientConfig, err error) {
	request := GetClientConfigRequest(config)

	response, err := serverapi.CallApiWithRequest(request, config)
	//response, err := serverapi.CallApi(http.MethodGet, "/clientconfig/", "", config)
	if err != nil {
		return clientConfig, nil
	}

	clientConfig, err = response2ClientConfig(response)
	if err != nil {
		return nil, err
	}

	return clientConfig, nil
}
