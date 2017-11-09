package client

import (
	"github.com/pkg/errors"
	"net/http"
)

type ClientConfig struct {
	GitUrlPattern string `json:"gitUrlPattern"`
}

func (api *ApiClient) GetClientConfig() (*ClientConfig, error) {
	endpoint := "/clientconfig/"

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var gc ClientConfig
	err = response.ParseFirstItem(&gc)
	if err != nil {
		return nil, errors.Wrap(err, "git config")
	}

	return &gc, nil
}
