package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiClient_Deploy(t *testing.T) {

	t.Run("Should successfully deploy applications", func(t *testing.T) {
		fileName := "deploy_paas_success_response"
		data := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Error(err)
				return
			}

			payload := `{"applicationIds":[{"environment":"boober-utv","application":"reference"}],"overrides":{},"deploy":true}`
			assert.JSONEq(t, payload, string(body))

			w.Write(data)
		}))
		defer ts.Close()

		applications := []string{"boober-utv/reference"}

		api := NewApiClient(ts.URL, "test", affiliation)
		deployPayload := NewDeployPayload(applications, make(map[string]json.RawMessage))
		deploys, err := api.Deploy(deployPayload)

		assert.NoError(t, err)
		assert.Len(t, deploys.Results, 1)
	})
}
