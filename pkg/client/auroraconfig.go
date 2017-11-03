package client

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AuroraConfig struct {
	Files    map[string]json.RawMessage `json:"files"`
	Versions map[string]string          `json:"versions"`
}

func NewAuroraConfig() *AuroraConfig {
	return &AuroraConfig{
		Files:    make(map[string]json.RawMessage),
		Versions: make(map[string]string),
	}
}

type auroraConfigFileNamesResponse struct {
	Response
	Items []string `json:"items"`
}

type auroraConfigResponse struct {
	Response
	Items []AuroraConfig `json:"items"`
}

func (api *ApiClient) GetFileNames() ([]string, *ErrorResponse) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/filenames", api.Affiliation)

	var res auroraConfigFileNamesResponse
	errorResponse, err := api.Do(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &res)
		return res, jErr
	})
	if err != nil {
		fmt.Println(err)
		return []string{}, errorResponse
	}

	return res.Items, nil
}

func (api *ApiClient) GetAuroraConfig() (*AuroraConfig, *ErrorResponse) {

	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)

	var acr auroraConfigResponse
	errorResponse, err := api.Do(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})
	if err != nil {
		fmt.Println(err)
		return nil, errorResponse
	}

	// TODO: Check for empty
	return &acr.Items[0], nil
}

// TODO: Return single AuroraConfig
func (api *ApiClient) SaveAuroraConfig(ac *AuroraConfig) ([]AuroraConfig, *ErrorResponse) {

	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)
	return api.putAuroraConfig(ac, endpoint)
}

func (api *ApiClient) ValidateAuroraConfig(ac *AuroraConfig) ([]AuroraConfig, *ErrorResponse) {

	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/validate", api.Affiliation)
	return api.putAuroraConfig(ac, endpoint)
}

func (api *ApiClient) putAuroraConfig(ac *AuroraConfig, endpoint string) ([]AuroraConfig, *ErrorResponse) {

	payload, err := json.Marshal(ac)
	if err != nil {
		logrus.Error("Failed to marshal AuroraConfig")
		return []AuroraConfig{}, nil
	}

	var acr auroraConfigResponse
	errorResponse, err := api.Do(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})
	if err != nil {
		fmt.Println(err)
		return []AuroraConfig{}, errorResponse
	}

	return acr.Items, nil
}
