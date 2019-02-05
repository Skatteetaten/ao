package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClient_GetClientConfig(t *testing.T) {
	t.Run("Should successfully get client config", func(t *testing.T) {
		fileName := "clientconfig_paas_success_response"
		data := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}))
		defer ts.Close()

		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		clientConfig, err := api.GetClientConfig()

		assert.NoError(t, err)
		assert.Equal(t, "file:///tmp/boober/%s", clientConfig.GitUrlPattern)
	})
}
