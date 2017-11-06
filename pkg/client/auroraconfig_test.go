package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func AuroraConfigSuccessResponseHandler(t *testing.T, responseFile string) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)

		assert.Contains(t, req.URL.Path, affiliation)

		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func AuroraConfigFailedResponseHandler(t *testing.T, responseFile string, status int) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)

		assert.Contains(t, req.URL.Path, affiliation)

		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func TestApi_GetAuroraConfig(t *testing.T) {

	t.Run("Successfully get AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		ac, errResponse := api.GetAuroraConfig()

		assert.Empty(t, errResponse)
		assert.Len(t, ac.Files, 13)
		assert.Len(t, ac.Versions, 13)
	})
}

func TestApiClient_GetFileNames(t *testing.T) {
	t.Run("Should get all filenames in AuroraConfig for a given affiliation", func(t *testing.T) {

		fileName := "filenames_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		fileNames, err := api.GetFileNames()

		assert.NoError(t, err)
		assert.Len(t, fileNames, 4)
	})
}

func TestApiClient_PutAuroraConfig(t *testing.T) {
	t.Run("Successfully validate and save AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)

		data := ReadTestFile("auroraconfig_paas_success_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}

		errResponse, err := api.ValidateAuroraConfig(&ac)

		assert.NoError(t, err)
		assert.Empty(t, errResponse)

		errResponse, err = api.SaveAuroraConfig(&ac)

		assert.NoError(t, err)
		assert.Empty(t, errResponse)
	})

	t.Run("Validation and save should fail when deploy type is illegal", func(t *testing.T) {
		fileName := "auroraconfig_paas_failed_validation_response"
		// TODO: This should not return error code 500
		ts := httptest.NewServer(AuroraConfigFailedResponseHandler(t, fileName, http.StatusInternalServerError))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		data := ReadTestFile("auroraconfig_paas_fail_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}

		errResponse, err := api.ValidateAuroraConfig(&ac)
		if errResponse == nil {
			t.Error("Expected errResponse to not be nil")
		}
		assert.NoError(t, err)
		assert.NotEmpty(t, errResponse)
		assert.Len(t, errResponse.IllegalFieldErrors, 1)

		errResponse, err = api.SaveAuroraConfig(&ac)
		assert.NoError(t, err)
		assert.NotEmpty(t, errResponse)
		assert.Len(t, errResponse.IllegalFieldErrors, 1)
	})
}
