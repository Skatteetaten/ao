package boober

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type AuroraConfig struct {
	Files    map[string]json.RawMessage `json:"files"`
	Versions map[string]string          `json:"versions"`
}

type auroraConfigFileNamesResponse struct {
	Response
	Items []string `json:"items"`
}

type auroraConfigResponse struct {
	Response
	Items []AuroraConfig `json:"items"`
}

func (api *BooberClient) GetFileNames() ([]string, *Validation) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/filenames", api.Affiliation)

	var res auroraConfigFileNamesResponse
	validation, err := api.Call(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &res)
		return res, jErr
	})
	if err != nil {
		fmt.Println(err)
		return []string{}, validation
	}

	return res.Items, validation
}

func (api *BooberClient) GetAuroraConfig() ([]AuroraConfig, *Validation) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)

	var acr auroraConfigResponse
	validation, err := api.Call(http.MethodGet, endpoint, nil, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})
	if err != nil {
		fmt.Println(err)
		return []AuroraConfig{}, validation
	}

	return acr.Items, validation
}

func (api *BooberClient) SaveAuroraConfig(ac *AuroraConfig) ([]AuroraConfig, *Validation) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig", api.Affiliation)
	return api.putAuroraConfig(ac, endpoint)
}

func (api *BooberClient) ValidateAuroraConfig(ac *AuroraConfig) ([]AuroraConfig, *Validation) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/validate", api.Affiliation)
	return api.putAuroraConfig(ac, endpoint)
}

func (api *BooberClient) putAuroraConfig(ac *AuroraConfig, endpoint string) ([]AuroraConfig, *Validation) {
	payload, err := json.Marshal(ac)
	if err != nil {
		logrus.Error("Failed to marshal AuroraConfig")
		return []AuroraConfig{}, nil
	}

	var acr auroraConfigResponse
	validation, err := api.Call(http.MethodPut, endpoint, payload, func(body []byte) (ResponseBody, error) {
		jErr := json.Unmarshal(body, &acr)
		return acr, jErr
	})
	if err != nil {
		fmt.Println(err)
		return []AuroraConfig{}, validation
	}

	return acr.Items, validation
}
