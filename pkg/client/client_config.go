package client

type ClientConfig struct {
	GitUrlPattern string `json:"gitUrlPattern"`
	ApiVersion    int    `json:"apiVersion"`
}

func (api *ApiClient) GetClientConfig() (*ClientConfig, error) {
	clientConfigGraphqlRequest := `{auroraApiMetadata{clientConfig{gitUrlPattern apiVersion}}}`
	type ClientConfigResponse struct {
		AuroraApiMetadata struct {
			ClientConfig ClientConfig
		}
	}

	var clientConfigResponse ClientConfigResponse
	if err := api.RunGraphQl(clientConfigGraphqlRequest, &clientConfigResponse); err != nil {
		return nil, err
	}
	gc := clientConfigResponse.AuroraApiMetadata.ClientConfig

	return &gc, nil
}
