package auroraconfig

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_jsonRemoveEntryRecursive_Do(t *testing.T) {
	t.Run("Should remove entry from normal JSON content", func(t *testing.T) {
		content := `{
		    "baseFile": "myapp.json",
			"cluster": "utv",
			"config": {
			    "MYAPP_SOME_KEY": "somevalue",
				"MYAPP_SOME_OTHER_KEY": "someothervalue",
                "MYAPP_KEYTOREMOVE": "sometrash"
			},
			"replicas": "1",
			"version": "1.2.3"
		}`
		pathParts := []string{"config", "MYAPP_KEYTOREMOVE"}
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = jsonRemoveEntryRecursive(&jsonContent, pathParts)
		assert.Nil(t, err)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.NotContains(t, changedjson, "MYAPP_KEYTOREMOVE")
		assert.NotContains(t, changedjson, "sometrash")
	})
	t.Run("Should fail when trying to remove non-existant entry", func(t *testing.T) {
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
		pathParts := []string{"config", "MYAPP_KEYTOREMOVE"}
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = jsonRemoveEntryRecursive(&jsonContent, pathParts)
		assert.NotNil(t, err)
		assert.Contains(t, "No such path in target JSON document", err.Error())
	})
	t.Run("Should fail when trying to remove named entry without full path", func(t *testing.T) {
		content := `{
		    "baseFile": "myapp.json",
			"cluster": "utv",
			"config": {
			    "MYAPP_SOME_KEY": "somevalue",
				"MYAPP_SOME_OTHER_KEY": "someothervalue",
                "MYAPP_KEYTOREMOVE": "sometrash"
			},
			"replicas": "1",
			"version": "1.2.3"
		}`
		pathParts := []string{"MYAPP_KEYTOREMOVE"}
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = jsonRemoveEntryRecursive(&jsonContent, pathParts)
		assert.NotNil(t, err)
		assert.Contains(t, "No such path in target JSON document", err.Error())

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "MYAPP_KEYTOREMOVE")
		assert.Contains(t, changedjson, "sometrash")
	})
	t.Run("Should remove subtree if it matches the key", func(t *testing.T) {
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
		pathParts := []string{"config"}
		var jsonContent map[string]interface{}
		err := json.Unmarshal([]byte(content), &jsonContent)
		assert.Nil(t, err)

		err = jsonRemoveEntryRecursive(&jsonContent, pathParts)
		assert.Nil(t, err)

		changedjsonbytearray, err := json.Marshal(jsonContent)
		assert.Nil(t, err)
		changedjson := string(changedjsonbytearray)
		assert.NotNil(t, changedjson)
		assert.NotContains(t, changedjson, "config")
		assert.NotContains(t, changedjson, "MYAPP_SOME_KEY")
		assert.NotContains(t, changedjson, "somevalue")
		assert.NotContains(t, changedjson, "MYAPP_SOME_OTHER_KEY")
		assert.NotContains(t, changedjson, "someothervalue")
		assert.Equal(t, 4, len(jsonContent))
	})
}

func Test_jsonSetOrCreateRecursive_Do(t *testing.T) {
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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
		assert.Nil(t, err)

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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
		assert.Nil(t, err)

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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
		assert.Nil(t, err)

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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
		assert.Nil(t, err)

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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
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

		err = jsonSetOrCreateRecursive(&jsonContent, pathParts, value)
		assert.NotNil(t, err)
		assert.Equal(t, "Path can not have numeric entries", err.Error())
	})
}
