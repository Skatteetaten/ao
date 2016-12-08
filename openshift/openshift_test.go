// Copyright © 2016 Skatteetaten <utvpaas@skatteetaten.no>
package openshift

import (
	"errors"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestPing(t *testing.T) {
	defer gock.Off()
	gock.InterceptClient(&client)

	testCases := []struct {
		statusCode int
		reachable  bool
		message    string
	}{
		{404, false, "Should not be reachable with error code not 200"},
		{200, true, "Should be reachable on code 200"},
	}
	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			url := "http://asdf.asdf"

			gock.New(url).
				Get("/").
				Reply(tc.statusCode)

			result := Ping(url)
			if result != tc.reachable {
				t.Error(tc.message)
			}
		})
	}

	t.Run("Should not be reachable with error", func(t *testing.T) {
		url := "http://asdf.asdf"
		gock.New(url).
			Get("/").
			ReplyError(errors.New("timed out"))

		result := Ping(url)
		if result {
			t.Error("Should not be reachable if error in http client")
		}

	})

}
