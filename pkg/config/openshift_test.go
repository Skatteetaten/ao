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
