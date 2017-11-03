package boober

import (
	"net/http"
	"encoding/json"
	"fmt"
)

type ClientConfig struct {
	GitUrlPattern    string `json:"gitUrlPattern"`
	OpenShiftCluster string `json:"openshiftCluster"`
	OpenShiftUrl     string `json:"openshiftUrl"`
}

type clientConfigResponse struct {
	Response
	Items []ClientConfig `json:"items"`
}

func (api *ApiClient) GetClientConfig() (ClientConfig, *ErrorResponse) {
	endpoint := "/clientconfig"

	var ccr clientConfigResponse
	errorResponse, err := api.Call(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &ccr)
		return ccr, jErr
	})
	if err != nil {
		fmt.Println(err)
		return ClientConfig{}, errorResponse
	}

	if len(ccr.Items) < 1 {
		errorResponse.SetMessage("No client config for affiliation " + api.Affiliation)
		return ClientConfig{}, errorResponse
	}

	return ccr.Items[0], nil
}
