package client

// Config specifies a client config
type Config struct {
	GitURLPattern string `json:"gitUrlPattern"`
	APIVersion    int    `json:"apiVersion"`
}

// GetClientConfig gets an client config via API calls
func (api *APIClient) GetClientConfig() (*Config, error) {
	clientConfigGraphqlRequest := `{auroraApiMetadata{clientConfig{gitUrlPattern apiVersion}}}`
	type ClientConfigResponse struct {
		AuroraAPIMetadata struct {
			ClientConfig Config
		}
	}

	var clientConfigResponse ClientConfigResponse
	if err := api.RunGraphQl(clientConfigGraphqlRequest, &clientConfigResponse); err != nil {
		return nil, err
	}
	gc := clientConfigResponse.AuroraAPIMetadata.ClientConfig

	return &gc, nil
}
