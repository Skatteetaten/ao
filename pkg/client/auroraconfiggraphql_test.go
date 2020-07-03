package client

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiClient_CreateAuroraConfigFile(t *testing.T) {
	t.Run("Should create a new aurora config file", func(t *testing.T) {
		fileName := "basic_auroraconfig"
		filecontent := ReadTestFile(fileName)
		responseFileName := "createauroraconfigfile_success_response"
		response := ReadTestFile(responseFileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "test", affiliation)
		createAuroraConfigFileResponse, err := api.CreateAuroraConfigFile("testconfig.json", filecontent)
		assert.NoError(t, err)
		assert.NotNil(t, createAuroraConfigFileResponse)
		assert.True(t, createAuroraConfigFileResponse.CreateAuroraConfigFile.Success)
		assert.Equal(t, "File successfully added", createAuroraConfigFileResponse.CreateAuroraConfigFile.Message)
	})
}
func TestApiClient_UpdateAuroraConfigFile(t *testing.T) {
	t.Run("Should update a aurora config file", func(t *testing.T) {
		fileName := "basic_auroraconfig"
		filecontent := ReadTestFile(fileName)
		acf := &auroraconfig.File{
			Name:     "testconfig.json",
			Contents: string(filecontent),
		}
		etag := "ETag"
		responseFileName := "updateauroraconfigfile_success_response"
		response := ReadTestFile(responseFileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "test", affiliation)
		updateAuroraConfigFileResponse, err := api.UpdateAuroraConfigFile(acf, etag)
		assert.NoError(t, err)
		assert.NotNil(t, updateAuroraConfigFileResponse)
		assert.True(t, updateAuroraConfigFileResponse.UpdateAuroraConfigFile.Success)
		assert.Equal(t, "File successfully updated", updateAuroraConfigFileResponse.UpdateAuroraConfigFile.Message)
	})
}
