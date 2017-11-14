package client

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type AuroraDeploySpec map[string]interface{}

func (api *ApiClient) GetAuroraDeploySpec(environment, application string) (AuroraDeploySpec, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/%s/%s", api.Affiliation, environment, application)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(response.Message)
	}

	var spec AuroraDeploySpec
	err = response.ParseFirstItem(&spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (api *ApiClient) GetAuroraDeploySpecFormatted(environment, application string) (string, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/%s/%s/formatted", api.Affiliation, environment, application)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", errors.New(response.Message)
	}

	var spec string
	err = response.ParseFirstItem(&spec)
	if err != nil {
		return "", err
	}

	return spec, nil
}
