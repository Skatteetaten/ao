package client

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var failedResponseText = `{
	"success": false,
	"message": "AuroraConfig contained errors",
	"items": [{
		"application": "foo",
		"environment": "bar",
		"messages": [{
			"type": "ILLEGAL",
			"message": "baz is not a legal value",
			"field": {
				"path": "/name/test",
				"value": "baz",
				"source": "about.json"
			}
		}]
	}],
	"count": 1
	}`

var responseText = `{
	"success": true,
	"message": "OK",
	"items": [{
		"files": {
			"about.json": "{}"
		},
		"versions": {
			"about.json": "tewt"
		}
	}],
	"count": 1
	}`

func TestResponse_ParseItemsWithErrors(t *testing.T) {

	var response Response
	err := json.Unmarshal([]byte(failedResponseText), &response)
	if err != nil {
		t.Error(err)
	}

	var acs []AuroraConfig
	err = response.ParseItems(&acs)
	if err == nil {
		t.Error("Expected response to contain errors")
	}
	fmt.Println(err)
}

func TestResponse_ParseItems(t *testing.T) {

	var response Response
	err := json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		t.Error(err)
	}

	var acs []AuroraConfig
	err = response.ParseItems(&acs)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(acs))

	for _, ac := range acs {
		_, found := ac.Files["about.json"]
		assert.Equal(t, true, found)
	}
}

func TestResponse_ParseFirstItem(t *testing.T) {

	var response Response
	err := json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		t.Error(err)
	}

	var ac AuroraConfig
	err = response.ParseFirstItem(&ac)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(ac.Files))
	assert.Equal(t, 1, len(ac.Versions))

	_, found := ac.Files["about.json"]
	assert.Equal(t, true, found)

	version, _ := ac.Versions["about.json"]
	assert.Equal(t, "tewt", version)
}
