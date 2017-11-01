package boober

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

func getTestServer(payload, body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusForbidden)
		res.Header().Set("Content-Type", "application/json")
		res.Write(body)
	}))
}

func TestApi_GetAuroraConfig(t *testing.T) {
	acBody := AuroraConfig{
		Files:    make(map[string]json.RawMessage),
		Versions: make(map[string]string),
	}

	response := auroraConfigResponse{
		Response: Response{
			Count: 1,
			Message: "OK",
			Success: true,
		},
		Items: []AuroraConfig{acBody},
	}

	body, _ := json.Marshal(response)

	ts := getTestServer(nil, body)
	defer ts.Close()

	api := NewBooberClient(ts.URL, "", "paas")
	ac, err := api.GetAuroraConfig()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 0, len(ac))
}
