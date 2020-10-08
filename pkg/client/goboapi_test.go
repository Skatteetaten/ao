package client

import (
	"github.com/skatteetaten/graphql"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func TestGoboApi(t *testing.T) {
	t.Run("Should work on normal graphql query", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := `{"data": {"someDataStructure": {"someData": ["testdata"]}}}}`

			data := []byte(response)
			w.Write(data)
		}))
		defer ts.Close()

		graphQlRequest := "{someDataStructure{someData}}"
		type SomeResponse struct {
			SomeDataStructure struct {
				SomeData []string
			}
		}
		var someResponse SomeResponse
		api := NewAPIClientDefaultRef(ts.URL, "test", affiliation)
		err := api.RunGraphQl(graphQlRequest, &someResponse)

		assert.NoError(t, err)
		assert.Equal(t, "testdata", someResponse.SomeDataStructure.SomeData[0])
	})

	t.Run("Should handle complex error message", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := `{"errors":[
                         {"message":"errors.message returned for first error","extensions":{"code":"CAN_NOT_FETCH_BY_ID",
                         "timestamp":"Fri Feb 9 14:33:09 UTC 2018","errors":[{"details":[{"type":"GENERIC"}],
                         "type":"APPLICATION"}]}},
                         {"message":"not returned for second error","extensions":{"errorMessage":"extensions.errorMessage returned for second error",
                         "code":"CAN_NOT_FETCH_BY_ID"}},
                         {"message":"not returned for third error",
                         "extensions":{"errorMessage":"not returned for third error","message":"not returned ever",
                         "errors":[{"details":[{"message":"extensions.errorMessage.errors.details.message returned for third error"}]},
                         {"details":[{"message":"twice"},{"message":"three times"}]}]}}]}`

			data := []byte(response)
			w.Write(data)
		}))
		defer ts.Close()

		graphQlRequest := "{someDataStructure{someData}}"
		type SomeResponse struct {
			SomeDataStructure struct {
				SomeData []string
			}
		}
		var someResponse SomeResponse
		api := NewAPIClientDefaultRef(ts.URL, "test", affiliation)
		err := api.RunGraphQl(graphQlRequest, &someResponse)

		assert.Error(t, err)
		assert.Equal(t, "errors.message returned for first error; extensions.errorMessage returned for second error; extensions.errorMessage.errors.details.message returned for third error; twice; three times", err.Error())
	})

	t.Run("Should work on normal graphql mutation", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			response := `{"data": {"someDataStructure": {"someData": ["testdata"]}}}}`

			data := []byte(response)
			w.Write(data)
		}))
		defer ts.Close()

		createStuffRequest := graphql.NewRequest(`mutation createStuff($newStuffInput: NewStuffInput!){	
			createStuff(input: $newStuffInput) {
    			message
    			success
  			}
		}`)
		createStuffRequest.Var("newStuffInput", "newstuff")
		type SomeResponse struct {
			SomeDataStructure struct {
				SomeData []string
			}
		}

		var someResponse SomeResponse
		api := NewAPIClientDefaultRef(ts.URL, "test", affiliation)
		err := api.RunGraphQlMutation(createStuffRequest, &someResponse)

		assert.NoError(t, err)
		assert.Equal(t, "testdata", someResponse.SomeDataStructure.SomeData[0])
	})
}
