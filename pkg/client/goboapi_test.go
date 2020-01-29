package client

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func TestGoboapi_Do(t *testing.T) {

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
		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		err := api.RunGraphQl(graphQlRequest, &someResponse)

		assert.NoError(t, err)
		assert.Equal(t, "testdata", someResponse.SomeDataStructure.SomeData[0])
	})
}
