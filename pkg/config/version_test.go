package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentVersionFromServer(t *testing.T) {

	// AO Version
	Version = "1.2.1"
	assert.Equal(t, "1.2.1", Version)

	// New AO version
	aoVersion := AOVersion{
		Version: "1.3.0",
	}

	t.Run("Should check for new version", func(t *testing.T) {
		response, err := json.Marshal(aoVersion)
		if err != nil {
			t.Error(err)
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		newVersion, err := GetCurrentVersionFromServer(ts.URL)
		assert.NoError(t, err)

		assert.Equal(t, "1.3.0", newVersion.Version)
		assert.Equal(t, true, newVersion.IsNewVersion())
	})
}

func TestGetNewAOClient(t *testing.T) {
	t.Run("Should get new ao version", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("1"))
		}))
		defer ts.Close()

		newAO, err := GetNewAOClient(ts.URL)
		assert.NoError(t, err)
		assert.NotEmpty(t, newAO)
	})
}
