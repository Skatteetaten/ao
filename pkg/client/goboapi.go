package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/graphql"
)

// RunGraphQl performs a GraphQl based API call
func (api *APIClient) RunGraphQl(graphQlRequest string, response interface{}) error {
	client := api.getGraphQlClient()
	req := api.newRequest(graphQlRequest)
	ctx := context.Background()

	if err := client.Run(ctx, req, response); err != nil {
		extractederr := extractGraphqlErrorMsg(err)
		return extractederr
	}
	return nil
}

func (api *APIClient) RunGraphQlMutation(graphQlRequest *graphql.Request, response interface{}) error {
	client := api.getGraphQlClient()
	ctx := context.Background()
	graphQlRequest.Header.Set("Cache-Control", "no-cache")
	graphQlRequest.Header.Add("Authorization", "Bearer "+api.Token)

	if err := client.Run(ctx, graphQlRequest, response); err != nil {
		extractederr := extractGraphqlErrorMsg(err)
		return extractederr
	}
	return nil
}

func (api *APIClient) getGraphQlClient() *graphql.Client {
	endpoint := fmt.Sprintf("%s/graphql", api.GoboHost)
	client := graphql.NewClient(endpoint)
	client.Log = func(logEntry string) { logrus.Debug(logEntry) }
	return client
}

func (api *APIClient) newRequest(graphqlRequest string) *graphql.Request {
	req := graphql.NewRequest(graphqlRequest)
	req.Header.Set("Cache-Control", "no-cache")
	return req
}

func extractGraphqlErrorMsg(err error) error {
	if err != nil {
		if graphqlErrors, ok := err.(graphql.Errors); ok {
			logrus.Debugf("extractGraphqlErrorMsg got %+v", graphqlErrors)
			// TODO
			// for _, e := range graphqlErrors {
			// extract message from each error
			// }
			return errors.New(fmt.Sprintf("errors: %+v", graphqlErrors))
		} else {
			logrus.Warnf("extractGraphqlErrorMsg got ordinary error (deprececated): %s", err)
			return err
		}
	} else {
		return errors.New("extractGraphqlErrorMsg got no error")
	}
}
