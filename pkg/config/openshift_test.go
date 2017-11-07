package config

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func TestAOConfig_InitClusters(t *testing.T) {

	ao := DefaultAOConfig
	ao.ClusterUrlPattern = "%s"
	ao.AvailableClusters = []string{}
	ao.PreferredAPIClusters = []string{}

	type TestCase struct {
		Name      string
		Reachable bool
	}

	testCases := []TestCase{
		{"prod", false},
		{"test", false},
		{"utv", true},
		{"qa", true},
	}

	testMap := make(map[string]TestCase)

	var testServers []*httptest.Server
	for _, test := range testCases {
		reachable := test.Reachable
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if reachable {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		}))
		testServers = append(testServers, ts)
		ao.AvailableClusters = append(ao.AvailableClusters, ts.URL)
		testMap[ts.URL] = test
	}

	noHost := "http://nohost:8080"
	ao.AvailableClusters = append(ao.AvailableClusters, noHost)
	testMap[noHost] = TestCase{
		Reachable: false,
		Name:      "no host",
	}

	// Test
	ao.InitClusters()
	for _, c := range ao.Clusters {
		test := testMap[c.Name]
		assert.Equal(t, test.Reachable, c.Reachable)

		// Since clusterUrlPattern is %s then name and url should be equal
		assert.Equal(t, c.Name, c.Url)
	}

	for _, t := range testServers {
		t.Close()
	}
}
