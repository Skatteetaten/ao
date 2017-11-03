package client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TESTING
// TODO: Test success get AuroraConfig
// TODO: Test failed put/validate AuroraConfig
// TODO: Test success deploy
// TODO: Test failed deploy

func AuroraConfigSuccessResponseHandler(affiliation string, t *testing.T) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {

		if !strings.Contains(req.RequestURI, "/affiliation") {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)

		uri := strings.TrimPrefix(req.RequestURI, "/affiliation/")
		uriAffiliation := strings.Split(uri, "/")[0]

		if affiliation != uriAffiliation {
			t.Errorf("Expected affiliation %s to be equal to %s", affiliation, uriAffiliation)
			return
		}

		fileName := fmt.Sprintf("auroraconfig_%s_success_response", affiliation)
		data := ReadTestFile(fileName)
		writer.Write(data)
	}
}

func TestApi_GetAuroraConfig(t *testing.T) {

	affiliation := "aurora"
	ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(affiliation, t))
	defer ts.Close()

	api := NewApiClient(ts.URL, "", affiliation)
	ac, errResponse := api.GetAuroraConfig()

	if errResponse != nil {
		t.Error("Expected ErrorResponse to be nil.")
	}

	assert.Equal(t, 90, len(ac.Files))

	if len(ac.Files) != len(ac.Versions) {
		t.Error("Expected Files and Version to have equal length.")
	}
}
