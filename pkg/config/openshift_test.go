package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.FatalLevel)
}

func TestCluster_HasValidToken(t *testing.T) {

	const (
		emptyToken   = ""
		invalidToken = "bar"
		validToken   = "foo"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") == "Bearer "+invalidToken {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	cases := []struct {
		Token    string
		Expected bool
	}{
		{emptyToken, false},
		{invalidToken, false},
		{validToken, true},
	}

	cluster := &Cluster{
		URL:  ts.URL,
		Name: "test",
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, cluster.IsValidToken(tc.Token))
	}
}

func TestGetToken(t *testing.T) {

	cases := []struct {
		RedirectPath  string
		ExpectedToken string
		ExpectedError string
	}{
		{`/?error=test&error_description=no host`, "", "test no host"},
		{"/#access_token=abc", "abc", ""},
		{"/", "", "token is empty"},
		{"/", "", "Not authorized"},
	}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, authenticationURLSuffix, req.URL.RequestURI())

			// No redirect when user is not authorized
			if tc.ExpectedError == "Not authorized" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			http.Redirect(w, req, tc.RedirectPath, http.StatusFound)
		}))

		token, err := GetToken(ts.URL, "", "")
		if err != nil {
			assert.EqualError(t, err, tc.ExpectedError)
		}

		assert.Equal(t, tc.ExpectedToken, token)

		ts.Close()
	}
}

func TestAOConfig_InitClusters(t *testing.T) {

	aoConfig := basicAOConfig
	aoConfig.ClusterURLPattern = "%s"
	aoConfig.BooberURLPattern = "%s"
	aoConfig.GoboURLPattern = "%s"
	aoConfig.AvailableClusters = []string{}
	aoConfig.PreferredAPIClusters = []string{}

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
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		testServers = append(testServers, ts)
		aoConfig.AvailableClusters = append(aoConfig.AvailableClusters, ts.URL)
		testMap[ts.URL] = test
	}

	noHost := "http://nohost:8080"
	aoConfig.AvailableClusters = append(aoConfig.AvailableClusters, noHost)
	testMap[noHost] = TestCase{
		Reachable: false,
		Name:      "no host",
	}

	// Test
	aoConfig.InitClusters()
	for _, c := range aoConfig.Clusters {
		test := testMap[c.Name]
		assert.Equal(t, test.Reachable, c.Reachable)

		// Since clusterUrlPattern is %s then name and url should be equal
		assert.Equal(t, c.Name, c.URL)
	}

	for _, t := range testServers {
		t.Close()
	}
}
