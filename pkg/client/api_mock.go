package client

import (
	"github.com/stretchr/testify/mock"
)

// APIClientMock is a base mock type
type APIClientMock struct {
	mock.Mock
}

// Do default mock implementation
func (api *APIClientMock) Do(method string, endpoint string, payload []byte) (*BooberResponse, error) {
	return nil, nil
}

// DoWithHeader default mock implementation
func (api *APIClientMock) DoWithHeader(method string, endpoint string, header map[string]string, payload []byte) (*ResponseBundle, error) {
	return nil, nil
}
