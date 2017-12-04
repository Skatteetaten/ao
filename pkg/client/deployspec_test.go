package client

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiClient_GetAuroraDeploySpec(t *testing.T) {
	t.Run("Should get aurora deploy spec", func(t *testing.T) {
		fileName := "deployspec_response"
		responseBody := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		spec, err := api.GetAuroraDeploySpec("aotest", "redis", true)
		assert.NoError(t, err)

		assert.Len(t, spec, 14)
	})
}

func TestApiClient_GetAuroraDeploySpecFormatted(t *testing.T) {
	t.Run("Should get formatted aurora deploy spec", func(t *testing.T) {
		fileName := "deployspec_formatted_response"
		responseBody := ReadTestFile(fileName)
		expected, err := ioutil.ReadFile("./test_files/deployspec_formatted.txt")
		if err != nil {
			panic(err)
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		spec, err := api.GetAuroraDeploySpecFormatted("aotest", "redis", true)
		assert.NoError(t, err)

		assert.Equal(t, string(expected), spec)
	})
}
