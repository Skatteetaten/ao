package boober

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/pkg/errors"
)

type ClientConfig struct {
	GitUrlPattern    string `json:"gitUrlPattern"`
	OpenShiftCluster string `json:"openshiftCluster"`
	OpenShiftUrl     string `json:"openshiftUrl"`
}

func (api *Api) GetClientConfig() (ClientConfig, error) {
	endpoint := "/clientconfig"

	var res struct {
		Response
		Items []ClientConfig `json:"items"`
	}

	err := api.WithRequest(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &res)
		return res, jErr
	})

	if err != nil {
		return ClientConfig{}, err
	}

	if len(res.Items) < 1 {
		return ClientConfig{}, errors.New("No client config for affiliation " + api.Affiliation)
	}

	return res.Items[0], nil
}
