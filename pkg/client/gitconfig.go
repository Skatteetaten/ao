package client

import (
	"github.com/pkg/errors"
	"net/http"
)

type GitConfig struct {
	GitUrlPattern string `json:"gitUrlPattern"`
}

func (api *ApiClient) GetClientConfig() (*GitConfig, error) {
	endpoint := "/clientconfig/"

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var gc GitConfig
	response.ParseFirstItem(&gc)
	if err != nil {
		return nil, errors.Wrap(err, "git config")
	}

	return &gc, nil
}
