package client

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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

func (api *ApiClient) GetFileNames() ([]string, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/filenames", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	err = response.ParseItems(&fileNames)
	if err != nil {
		return nil, err
	}

	return fileNames, err
}

func (api *ApiClient) GetAuroraConfig() (*AuroraConfig, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var ac AuroraConfig
	err = response.ParseFirstItem(&ac)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config")
	}

	return &ac, nil
}

func (api *ApiClient) PutAuroraConfig(endpoint string, ac *AuroraConfig) (*ErrorResponse, error) {

	payload, err := json.Marshal(ac)
	if err != nil {
		return nil, err
	}

	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return response.ToErrorResponse()
	}

	return nil, nil
}

func (api *ApiClient) SaveAuroraConfig(ac *AuroraConfig) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)
	return api.PutAuroraConfig(endpoint, ac)
}

func (api *ApiClient) ValidateAuroraConfig(ac *AuroraConfig) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/validate", api.Affiliation)
	return api.PutAuroraConfig(endpoint, ac)
}
