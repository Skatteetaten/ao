package client

import (
	"fmt"
	"net/http"
	"net/url"

	"ao/pkg/deploymentspec"
)

type DeploySpecClient interface {
	Doer
	GetAuroraDeploySpec(applications []string, defaults bool) ([]deploymentspec.DeploymentSpec, error)
	GetAuroraDeploySpecFormatted(environment, application string, defaults bool) (string, error)
}

func (api *ApiClient) GetAuroraDeploySpec(applications []string, defaults bool) ([]deploymentspec.DeploymentSpec, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/?", api.Affiliation)
	queries := buildDeploySpecQueries(applications, defaults)

	adsCh := make(chan []deploymentspec.DeploymentSpec)
	errCh := make(chan error)
	for _, q := range queries {
		go func(path, query string) {
			response, err := api.Do(http.MethodGet, endpoint+query, nil)
			if err != nil {
				errCh <- err
				return
			}

			var specs []deploymentspec.DeploymentSpec
			err = response.ParseItems(&specs)
			if err != nil {
				errCh <- err
			} else {
				adsCh <- specs
			}
		}(endpoint, q)
	}

	var allSpecs []deploymentspec.DeploymentSpec
	for i := 0; i < len(queries); i++ {
		select {
		case err := <-errCh:
			return nil, err
		case spec := <-adsCh:
			allSpecs = append(allSpecs, spec...)
		}
	}

	// Must copy elements to array of interfaces.
	// TODO: Find a way to avoid having to do this.
	deploySpecs := make([]deploymentspec.DeploymentSpec, len(allSpecs))
	for i, spec := range allSpecs {
		deploySpecs[i] = deploymentspec.DeploymentSpec(spec)
	}

	return deploySpecs, nil
}

func buildDeploySpecQueries(applications []string, defaults bool) []string {
	const maxQueryLength = 3500
	var queries []string

	v := url.Values{}
	for _, app := range applications {
		v.Add("aid", app)
		if len(v.Encode()) >= maxQueryLength {
			if !defaults {
				v.Add("includeDefaults", "false")
			}
			queries = append(queries, v.Encode())
			v = url.Values{}
		}
	}
	if !defaults {
		v.Add("includeDefaults", "false")
	}

	return append(queries, v.Encode())
}

func (api *ApiClient) GetAuroraDeploySpecFormatted(environment, application string, defaults bool) (string, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/%s/%s/formatted", api.Affiliation, environment, application)
	if !defaults {
		endpoint += "?includeDefaults=false"
	}

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	var spec string
	err = response.ParseFirstItem(&spec)
	if err != nil {
		return "", err
	}

	return spec, nil
}
