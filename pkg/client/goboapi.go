package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/graphql"
	"strings"
)

// RunGraphQl performs a GraphQl based API call
func (api *APIClient) RunGraphQl(graphQlRequest string, response interface{}) error {
	client := api.getGraphQlClient()
	req := api.newRequest(graphQlRequest)
	ctx := context.Background()

	if err := client.Run(ctx, req, response); err != nil {
		extractederr := extractGraphqlErrorMsgs(err)
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
		extractederr := extractGraphqlErrorMsgs(err)
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

func extractGraphqlErrorMsgs(errorsInput error) error {
	if errorsInput != nil {
		if graphqlErrors, ok := errorsInput.(graphql.Errors); ok {
			logrus.Debugf("extractGraphqlErrorMsg got %+v", graphqlErrors)
			errorMsgs := make([]string, len(graphqlErrors))
			for i, e := range graphqlErrors {
				errorMsgs[i] = getPrioritizedErrMsg(e)
			}
			return errors.New(strings.Join(errorMsgs, "; "))
		} else {
			logrus.Warnf("extractGraphqlErrorMsg got ordinary error (deprececated): %s", errorsInput)
			return errorsInput
		}
	} else {
		return errors.New("extractGraphqlErrorMsg got no error")
	}
}

// Extract message from error by prioritised level
// 1. error.extensions.errors[...].details[...].message
// 2. error.extensions.errorMessage
// 3. error.message (default)
func getPrioritizedErrMsg(graphqlErr graphql.Error) string {
	extensions, parseError := parseExtensions(graphqlErr.Extensions)
	if parseError == nil && extensions != nil {
		extErrorMsg := extensions.getErrorMessage()
		if extErrorMsg != "" {
			return extErrorMsg
		}
	}
	// 3. error.message (default)
	if graphqlErr.Message == "" {
		if parseError != nil {
			return "got unparseable extended error response from server"
		}
		return "got unspecified error"
	}
	return graphqlErr.Message
}

func parseExtensions(unparsedExtensions map[string]interface{}) (*extensions, error) {
	if unparsedExtensions != nil {
		extensions := extensions{}
		jsonExtensions, err := json.Marshal(unparsedExtensions)
		if err != nil {
			logrus.Warnf("Could not get json from %+v", unparsedExtensions)
			return nil, err
		}
		if err := json.Unmarshal(jsonExtensions, &extensions); err != nil {
			logrus.Warnf("Could not parse extensions from %s", jsonExtensions)
			return nil, err
		}
		return &extensions, nil
	}
	return nil, nil
}

// Extension extends the standard GraphQl error structure with more application specific details
type extensions struct {
	Code           string
	ErrorMessage   string
	SourceSystem   string
	Message        string
	ExtErrors      []extError `json:"errors"`
	Classification string
}

func (ext extensions) getErrorMessage() string {
	// 1. error.extensions.errors[...].details[...].message
	if ext.ExtErrors != nil && len(ext.ExtErrors) > 0 {
		detailMsgs := ext.getDetailsErrorMessages()
		if len(detailMsgs) > 0 {
			return strings.Join(detailMsgs, "; ")
		}
	}
	// 2. error.extensions.errorMessage
	if ext.ErrorMessage != "" {
		return ext.ErrorMessage
	}
	return ""
}

func (ext extensions) getDetailsErrorMessages() []string {
	detailMsgs := make([]string, 0)
	for _, extError := range ext.ExtErrors {
		detailMsgs = extError.appendDetailsMessages(detailMsgs)
	}
	return detailMsgs
}

type extError struct {
	Application string
	Environment string
	Details     []detail
	Type        string
}

func (extError extError) appendDetailsMessages(detailMsgs []string) []string {
	if extError.Details != nil && len(extError.Details) > 0 {
		for _, detail := range extError.Details {
			detailMsgs = detail.appendMessage(detailMsgs)
		}
	}
	return detailMsgs
}

type detail struct {
	Type    string
	Message string
}

func (detail detail) appendMessage(detailMsgs []string) []string {
	if detail.Message != "" {
		detailMsgs = append(detailMsgs, detail.Message)
	}
	return detailMsgs
}
