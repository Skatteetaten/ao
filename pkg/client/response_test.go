package client

import (
	"encoding/json"
	"testing"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/stretchr/testify/assert"
)

var failedResponseText = `{
	"success": false,
	"message": "AuroraConfig contained errors",
	"items": [{
		"application": "foo",
		"environment": "bar",
		"details": [{
			"type": "ILLEGAL",
			"message": "baz is not a legal value",
			"field": {
				"handler": {
				  "path": "/name/test"
				},
				"source": {
				  "configName": "about.json",
				  "contents": "",
				  "name": "about.json",
				  "override": false
				},
				"value": "baz"
		}}, {
		  "type": "INVALID",
		  "message": "/asdlkjf is not a valid config field pointer",
		  "field": {
				"handler": {
				  "path": "/asdlkjf"
				},
				"source": {
				  "configName": "boober-utv/reference.json",
				  "contents": "",
				  "name": "boober-utv/reference.json",
				  "override": false
				},
				"value": null
		  }}, {
		  "type": "MISSING",
		  "message": "Name must be alphanumeric and no more than 40 characters",
		  "field": {
				"handler": {
				  "path": "/name"
				},
				"source": {
				  "configName": "boober-utv/reference.json",
				  "contents": "",
				  "name": "boober-utv/reference.json",
				  "override": false
				},
				"value": null
		}},{
		  "type": "GENERIC",
		  "message": "Vault random does not exists"
		}]
	}],
	"count": 3
}`

var responseText = `{
	"success": true,
	"message": "OK",
	"items": [{
		"name": "aurora",
		"files": [{
			"name": "about.json",
			"contents": "{}"
		}]
	}],
	"count": 1
	}`

func TestResponse_ParseItemsWithErrors(t *testing.T) {

	var response BooberResponse
	err := json.Unmarshal([]byte(failedResponseText), &response)
	if err != nil {
		t.Error(err)
	}

	var acs []auroraconfig.AuroraConfig
	err = response.ParseItems(&acs)
	if err == nil {
		t.Error("Expected response to contain errors")
	}
}

func TestResponse_ParseItems(t *testing.T) {

	var response BooberResponse
	err := json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		t.Error(err)
	}

	var acs []auroraconfig.AuroraConfig
	err = response.ParseItems(&acs)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(acs))

	for _, ac := range acs {
		assert.Len(t, ac.Files, 1)
		assert.Equal(t, "about.json", ac.Files[0].Name)
	}
}

func TestResponse_ParseFirstItem(t *testing.T) {

	var response BooberResponse
	err := json.Unmarshal([]byte(responseText), &response)
	if err != nil {
		t.Error(err)
	}

	var ac auroraconfig.AuroraConfig
	err = response.ParseFirstItem(&ac)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(ac.Files))
	assert.Equal(t, "about.json", ac.Files[0].Name)
}

func TestResponse_ToErrorResponse(t *testing.T) {

	var response BooberResponse
	err := json.Unmarshal([]byte(failedResponseText), &response)
	if err != nil {
		t.Error(err)
	}

	errorResponse, err := response.toErrorResponse()
	assert.NoError(t, err)

	assert.Len(t, errorResponse.GenericErrors, 1)
	assert.Len(t, errorResponse.IllegalFieldErrors, 1)
	assert.Len(t, errorResponse.InvalidFieldErrors, 1)
	assert.Len(t, errorResponse.MissingFieldErrors, 1)

	assert.Len(t, errorResponse.getAllErrors(), 4)
}
