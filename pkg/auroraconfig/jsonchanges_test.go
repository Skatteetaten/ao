package auroraconfig

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetValue_Do(t *testing.T) {
	t.Run("Should set value in AuroraConfigFile (happy test)", func(t *testing.T) {
		content := `{
		    "baseFile": "myapp.json"
		}`
		auroraConfigFile := AuroraConfigFile{
			Name:     "myconfigfile.json",
			Contents: content,
		}
		path := "/config/MYAPP_NEW_KEY"
		value := "newValue"

		SetValue(&auroraConfigFile, path, value)

		changedjson := auroraConfigFile.Contents
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "MYAPP_NEW_KEY")
		assert.Contains(t, changedjson, "newValue")
		assert.Equal(t, "{\n  \"baseFile\": \"myapp.json\",\n  \"config\": {\n    \"MYAPP_NEW_KEY\": \"newValue\"\n  }\n}", changedjson)
	})
}

func TestSetOrCreate_Do(t *testing.T) {
	t.Run("Should set value on normal JSON content", func(t *testing.T) {
		content := `{
		    "baseFile": "myapp.json",
			"cluster": "utv",
			"config": {
			    "MYAPP_SOME_KEY": "somevalue",
				"MYAPP_SOME_OTHER_KEY": "someothervalue"
			},
			"replicas": "1",
			"version": "1.2.3"
		}`
		pathParts := []string{"config", "MYAPP_NEW_KEY"}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		setOrCreate(&jsonContent, pathParts, value)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "MYAPP_NEW_KEY")
		assert.Contains(t, changedjson, "newValue")
	})

	t.Run("Should set value on minimal JSON content", func(t *testing.T) {
		content := `{}`
		pathParts := []string{"MYAPP_NEW_KEY"}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		setOrCreate(&jsonContent, pathParts, value)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "MYAPP_NEW_KEY")
		assert.Contains(t, changedjson, "newValue")
		assert.Equal(t, `{"MYAPP_NEW_KEY":"newValue"}`, changedjson)
	})

	t.Run("Should set new value on new multi level path", func(t *testing.T) {
		content := `{"baseFile": "myapp.json"}`
		pathParts := []string{"first", "second", "MYAPP_NEW_KEY"}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		setOrCreate(&jsonContent, pathParts, value)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.Equal(t, `{"baseFile":"myapp.json","first":{"second":{"MYAPP_NEW_KEY":"newValue"}}}`, changedjson)
	})

	t.Run("Should replace existing value", func(t *testing.T) {
		content := `{"baseFile": "myapp.json", "config": {"MYAPP_SOME_KEY": "somevalue"}}`
		pathParts := []string{"config", "MYAPP_SOME_KEY"}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		setOrCreate(&jsonContent, pathParts, value)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "newValue")
		assert.NotContains(t, changedjson, "somevalue")
		assert.Equal(t, `{"baseFile":"myapp.json","config":{"MYAPP_SOME_KEY":"newValue"}}`, changedjson)
	})

	t.Run("Should fail with empty path", func(t *testing.T) {
		content := `{"baseFile": "myapp.json", "config": {"MYAPP_SOME_KEY": "somevalue"}}`
		pathParts := []string{}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = setOrCreate(&jsonContent, pathParts, value)
		assert.NotNil(t, err)
		assert.Equal(t, "Path can not be empty", err.Error())
	})

	t.Run("Should fail on numeric entry in path", func(t *testing.T) {
		content := `{"baseFile": "myapp.json", "config": {"MYAPP_SOME_KEY": "somevalue"}}`
		pathParts := []string{"config", "270"}
		value := "newValue"
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = setOrCreate(&jsonContent, pathParts, value)
		assert.NotNil(t, err)
		assert.Equal(t, "Path can not have numeric entries", err.Error())
	})
}

func TestGetPartsPath_Do(t *testing.T) {

	t.Run("Should handle normal path: /someattribute", func(t *testing.T) {

		pathParts := getPathParts("/someattribute")

		assert.NotNil(t, pathParts)
		assert.Equal(t, 1, len(pathParts))
		assert.Equal(t, "someattribute", pathParts[0])
	})

	t.Run("Should return correct response on multiple step path: /one/number2/three/four/FIVE/Six/S_E_V_E_N", func(t *testing.T) {

		pathParts := getPathParts("/one/number2/three/four/FIVE/Six/S_E_V_E_N")

		assert.NotNil(t, pathParts)
		assert.Equal(t, 7, len(pathParts))
		assert.Equal(t, "one", pathParts[0])
		assert.Equal(t, "number2", pathParts[1])
		assert.Equal(t, "three", pathParts[2])
		assert.Equal(t, "four", pathParts[3])
		assert.Equal(t, "FIVE", pathParts[4])
		assert.Equal(t, "Six", pathParts[5])
		assert.Equal(t, "S_E_V_E_N", pathParts[6])
	})

	t.Run("Should handle missing / at start of path: someattribute", func(t *testing.T) {

		pathParts := getPathParts("someattribute")

		assert.NotNil(t, pathParts)
		assert.Equal(t, 1, len(pathParts))
		assert.Equal(t, "someattribute", pathParts[0])
	})

	t.Run("Should handle / at end of path: someattribute/", func(t *testing.T) {

		pathParts := getPathParts("someattribute/")

		assert.NotNil(t, pathParts)
		assert.Equal(t, 1, len(pathParts))
		assert.Equal(t, "someattribute", pathParts[0])
	})

	t.Run("Should return empty response on empty path", func(t *testing.T) {

		pathParts := getPathParts("")

		assert.Nil(t, pathParts)
		assert.Equal(t, 0, len(pathParts))
	})

	t.Run("Should return empty response on rooty path: /", func(t *testing.T) {

		pathParts := getPathParts("/")

		assert.NotNil(t, pathParts)
		assert.Equal(t, 0, len(pathParts))
	})
}
