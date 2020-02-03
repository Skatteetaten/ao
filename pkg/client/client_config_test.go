package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClient_GetClientConfig(t *testing.T) {
	t.Run("Should successfully get client config", func(t *testing.T) {
		fileName := "clientconfig_graphql_success_response"
		data := ReadTestFile(fileName)
		var calls int

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			calls++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}))
		defer ts.Close()

		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		clientConfig, err := api.GetClientConfig()

		assert.Equal(t, 1, calls)
		assert.NoError(t, err)
		assert.Equal(t, "file:///tmp/boober/%s", clientConfig.GitUrlPattern)
		assert.Equal(t, 2, clientConfig.ApiVersion)
	})
}
