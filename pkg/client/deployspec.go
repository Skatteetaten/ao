package client

import (
	"fmt"
	"net/http"
	"strings"
)

type AuroraDeploySpec map[string]interface{}

func (a AuroraDeploySpec) Value(jsonPointer string) interface{} {
	return a.get(jsonPointer + "/value")
}

func (a AuroraDeploySpec) get(jsonPointer string) interface{} {
	pointers := strings.Fields(strings.Replace(jsonPointer, "/", " ", -1))
	current := a
	for i, pointer := range pointers {
		isLast := i == len(pointers)-1
		if next, ok := current[pointer]; ok && isLast {
			return next
		} else if ok && !isLast {
			current = next.(map[string]interface{})
		} else {
			return "-"
		}
	}
	return "-"
}

func (api *ApiClient) GetAuroraDeploySpec(environment, application string, defaults bool) (AuroraDeploySpec, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/%s/%s", api.Affiliation, environment, application)
	if !defaults {
		endpoint += "?includeDefaults=false"
	}

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var spec AuroraDeploySpec
	err = response.ParseFirstItem(&spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
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

func (api *ApiClient) GetAuroraDeploySpecs(applications []string) []AuroraDeploySpec {
	specCh := make(chan AuroraDeploySpec)
	errorCh := make(chan error)
	for _, app := range applications {
		go func(id string) {
			split := strings.Split(id, "/")
			spec, err := api.GetAuroraDeploySpec(split[0], split[1], true)
			if err != nil {
				errorCh <- err
			} else {
				specCh <- spec
			}
		}(app)
	}

	var specs []AuroraDeploySpec
	for i := 0; i < len(applications); i++ {
		select {
		case err := <-errorCh:
			fmt.Println(err)
		case spec := <-specCh:
			specs = append(specs, spec)
		}
	}

	return specs
}
