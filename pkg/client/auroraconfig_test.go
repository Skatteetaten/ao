package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func AuroraConfigSuccessResponseHandler(responseFile string) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)

		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func AuroraConfigFailedResponseHandler(responseFile string, status int) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)

		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func TestApi_GetAuroraConfig(t *testing.T) {

	t.Run("Successfully get AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "paas")
		ac, errResponse := api.GetAuroraConfig()

		if errResponse != nil {
			t.Error("Expected ErrorResponse to be nil.")
		}

		assert.Equal(t, 13, len(ac.Files))

		if len(ac.Files) != len(ac.Versions) {
			t.Error("Expected Files and Version to have equal length.")
		}
	})
}

func TestApiClient_PutAuroraConfig(t *testing.T) {
	t.Run("Successfully validate AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "paas")

		data := ReadTestFile("auroraconfig_paas_success_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}

		errResponse, err := api.ValidateAuroraConfig(&ac)
		if err != nil {
			t.Error("Expected error to be nil.")
		}

		if errResponse != nil {
			t.Error("Expected ErrorResponse to be nil.")
		}
	})

	t.Run("Validation should fail when deploy type is illegal", func(t *testing.T) {
		fileName := "auroraconfig_paas_failed_validation_response"
		// TODO: This should not return error code 500
		ts := httptest.NewServer(AuroraConfigFailedResponseHandler(fileName, http.StatusInternalServerError))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "paas")
		data := ReadTestFile("auroraconfig_paas_fail_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}

		errResponse, err := api.ValidateAuroraConfig(&ac)
		if err != nil {
			t.Error("Expected error to be nil.")
		}

		if errResponse == nil {
			t.Error("Expected ErrorResponse to not be nil.")
		}

		assert.Equal(t, 1, len(errResponse.IllegalFieldErrors))
	})
}
