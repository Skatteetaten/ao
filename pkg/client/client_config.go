package client

import (
	"strconv"
)

type ClientConfig struct {
	GitUrlPattern string `json:"gitUrlPattern"`
	ApiVersion    int    `json:"apiVersion"`
}

func (api *ApiClient) GetClientConfig() (*ClientConfig, error) {
	clientConfigGraphqlRequest := `{auroraApiMetadata{clientConfig{gitUrlPattern apiVersion}}}`
	type ClientConfigResponse struct {
		AuroraApiMetadata struct {
			ClientConfig struct {
				GitUrlPattern string `json:"gitUrlPattern"`
				ApiVersion    string `json:"apiVersion"`
			}
		}
	}

	var clientConfigResponse ClientConfigResponse
	if err := api.RunGraphQl(clientConfigGraphqlRequest, &clientConfigResponse); err != nil {
		return nil, err
	}

	// Need to convert to int for reuse in existing domain logic
	apiVersion, err := strconv.Atoi(clientConfigResponse.AuroraApiMetadata.ClientConfig.ApiVersion)
	if err != nil {
		return nil, err
	}
	gc := ClientConfig{clientConfigResponse.AuroraApiMetadata.ClientConfig.GitUrlPattern, apiVersion}

	return &gc, nil
}
