package boober

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type AuroraConfig struct {
	Files    map[string]json.RawMessage `json:"files"`
	Versions map[string]string          `json:"versions"`
}

type AuroraConfigFileNamesResponse struct {
	Response
	Items []string `json:"items"`
}

type AuroraConfigResponse struct {
	Response
	Items []AuroraConfig `json:"items"`
}

func (api *Api) GetFileNames() ([]string, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/filenames", api.Affiliation)

	var res AuroraConfigFileNamesResponse
	err := api.WithRequest(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &res)
		return res, jErr
	})

	return res.Items, err
}

func (api *Api) GetAuroraConfig() ([]AuroraConfig, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)

	var acr AuroraConfigResponse
	err := api.WithRequest(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})

	return acr.Items, err
}

func (api *Api) PutAuroraConfig(ac AuroraConfig) ([]AuroraConfig, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)

	payload, err := json.Marshal(ac)
	if err != nil {
		return []AuroraConfig{}, err
	}

	var acr AuroraConfigResponse
	err = api.WithRequest(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})

	return acr.Items, err
}

func (api *Api) ValidateAuroraConfig(ac AuroraConfig) ([]AuroraConfig, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/validate", api.Affiliation)

	payload, err := json.Marshal(ac)
	if err != nil {
		return []AuroraConfig{}, err
	}

	var acr AuroraConfigResponse
	err = api.WithRequest(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})

	return acr.Items, err
}
