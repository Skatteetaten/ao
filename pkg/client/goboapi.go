package client

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
)

func (api *APIClient) RunGraphQl(graphQlRequest string, response interface{}) error {
	client := api.getGraphQlClient()
	req := api.newRequest(graphQlRequest)
	ctx := context.Background()

	if err := client.Run(ctx, req, response); err != nil {
		return err
	}
	return nil
}

func (api *APIClient) getGraphQlClient() *graphql.Client {
	endpoint := fmt.Sprintf("%s/graphql", api.GoboHost)
	client := graphql.NewClient(endpoint)
	return client
}

func (api *APIClient) newRequest(graphqlRequest string) *graphql.Request {
	req := graphql.NewRequest(graphqlRequest)
	req.Header.Set("Cache-Control", "no-cache")
	return req
}
