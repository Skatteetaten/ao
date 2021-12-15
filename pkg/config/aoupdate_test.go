package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAOConfig_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	aoConfig := &AOConfig{
		Clusters:                make(map[string]*Cluster),
		AvailableClusters:       []string{"test"},
		AvailableUpdateClusters: []string{"test"},
	}
	aoConfig.Clusters["test"] = &Cluster{
		Name:      "test",
		Reachable: true,
		UpdateURL: ts.URL,
	}

	url, err := aoConfig.getUpdateURL()

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s", ts.URL), url)
}
