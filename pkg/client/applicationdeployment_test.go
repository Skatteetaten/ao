package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClient_Deploy(t *testing.T) {

	t.Run("Should successfully deploy applications", func(t *testing.T) {
		fileName := "deploy_paas_success_response"
		response := ReadTestFile(fileName)

		expectedPayload := `{"applicationDeploymentRefs":[{"environment":"boober-utv","application":"reference"}],"overrides":{},"deploy":true}`

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Error(err)
				return
			}

			assert.JSONEq(t, expectedPayload, string(body))
			w.Write(response)
		}))
		defer ts.Close()

		applications := []string{"boober-utv/reference"}

		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		deployPayload := NewDeployPayload(applications, make(map[string]string))
		deploys, err := api.Deploy(deployPayload)

		assert.NoError(t, err)
		assert.Len(t, deploys.Results, 1)
	})
}

func TestApiClient_Delete(t *testing.T) {

	t.Run("Should successfully delete applications", func(t *testing.T) {
		response := ReadTestFile("delete_paas_success_response")
		expectedPayload := `{"applicationRefs":[{"namespace":"foo-dev","name":"bar"}]}`

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Error(err)
				return
			}

			assert.JSONEq(t, expectedPayload, string(body))
			w.Write(response)
		}))
		defer ts.Close()

		applications := []ApplicationRef{*NewApplicationRef("foo-dev", "bar")}

		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		deletePayload := NewDeletePayload(applications)
		deletes, err := api.Delete(deletePayload)

		assert.NoError(t, err)
		assert.Len(t, deletes.Results, 1)
	})
}
