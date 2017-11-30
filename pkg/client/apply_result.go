package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *ApiClient) GetApplyResult(deployId string) (string, error) {
	endpoint := fmt.Sprintf("/apply-result/%s/%s", api.Affiliation, deployId)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	var result json.RawMessage
	err = response.ParseFirstItem(&result)
	if err != nil {
		return "", err
	}

	applyResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(applyResult), nil
}
