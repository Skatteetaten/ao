package client

import (
	"fmt"
	"net/http"
	"net/url"
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

func (api *ApiClient) GetAuroraDeploySpec(applications []string, defaults bool) ([]AuroraDeploySpec, error) {
	endpoint := fmt.Sprintf("/auroradeployspec/%s/?", api.Affiliation)
	queries := buildDeploySpecQueries(applications, defaults)

	adsCh := make(chan []AuroraDeploySpec)
	errCh := make(chan error)
	for _, q := range queries {
		go func(path, query string) {
			response, err := api.Do(http.MethodGet, endpoint+query, nil)
			if err != nil {
				errCh <- err
				return
			}

			var specs []AuroraDeploySpec
			err = response.ParseItems(&specs)
			if err != nil {
				errCh <- err
			} else {
				adsCh <- specs
			}
		}(endpoint, q)
	}

	var allSpecs []AuroraDeploySpec
	for i := 0; i < len(queries); i++ {
		select {
		case err := <-errCh:
			return nil, err
		case spec := <-adsCh:
			allSpecs = append(allSpecs, spec...)
		}
	}

	return allSpecs, nil
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
