package client

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiClient_GetApplyResult(t *testing.T) {
	t.Run("Should successfully get apply result", func(t *testing.T) {

		deployId := "acba3"

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			expectedPath := fmt.Sprintf("/v1/apply-result/%s/%s", affiliation, deployId)
			assert.Equal(t, expectedPath, req.URL.Path)

			// Not a real apply result, just testing indenting from items and first item is received
			response := `{"success": true, "message": "OK", "items": [{"deploy": "failed"}], "count": 0}`
			w.Write([]byte(response))
		}))
		defer ts.Close()

		api := NewApiClientDefaultRef(ts.URL, "test", affiliation)
		result, err := api.GetApplyResult(deployId)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "{\n  \"deploy\": \"failed\"\n}", result)
	})
}
