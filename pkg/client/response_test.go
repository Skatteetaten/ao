package client

import (
	"encoding/json"
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
			}}, {
		  "type": "INVALID",
		  "message": "/asdlkjf is not a valid config field pointer",
		  "field": {
			"path": "/asdlkjf",
			"value": "",
			"source": "boober-utv/reference.json"
		  }}, {
		  "type": "MISSING",
		  "message": "Name must be alphanumeric and no more than 40 characters",
		  "field": {
			"path": "/name",
			"value": "",
			"source": "Unknown"
		  }}]
	}],
	"count": 3
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

func TestResponse_ToErrorResponse(t *testing.T) {

	var response Response
	err := json.Unmarshal([]byte(failedResponseText), &response)
	if err != nil {
		t.Error(err)
	}

	errorResponse, err := response.ToErrorResponse()
	assert.NoError(t, err)

	assert.Len(t, errorResponse.IllegalFieldErrors, 1)
	assert.Len(t, errorResponse.InvalidFieldErrors, 1)
	assert.Len(t, errorResponse.MissingFieldErrors, 1)

	assert.Len(t, errorResponse.GetAllErrors(), 3)
}

func TestErrorResponse_SetMessage(t *testing.T) {

	errorResponse := &ErrorResponse{
		UniqueErrors: make(map[string]bool),
	}

	assert.Equal(t, false, errorResponse.ContainsError)
	errorResponse.SetMessage("Failed")
	assert.Equal(t, true, errorResponse.ContainsError)
	assert.Equal(t, "Failed\n", errorResponse.String())

}
