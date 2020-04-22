package auroraconfig

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func Test_yamlRemoveEntryRecursive_Do(t *testing.T) {
	t.Run("Should remove entry from normal YAML content", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
cluster: utv
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
  MYAPP_KEYTOREMOVE: sometrash
replicas: '1'
version: 1.2.3
`
		pathParts := []string{"config", "MYAPP_KEYTOREMOVE"}
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlRemoveEntryRecursive(&yamlContent, pathParts)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.NotContains(t, changedyaml, "MYAPP_KEYTOREMOVE")
		assert.NotContains(t, changedyaml, "sometrash")
	})
	t.Run("Should fail when trying to remove non-existant entry", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
cluster: utv
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: '1'
version: 1.2.3
`
		pathParts := []string{"config", "MYAPP_KEYTOREMOVE"}
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlRemoveEntryRecursive(&yamlContent, pathParts)
		assert.NotNil(t, err)
		assert.Contains(t, "No such path in target YAML document", err.Error())
	})
	t.Run("Should fail when trying to remove named entry without full path", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
cluster: utv
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
  MYAPP_KEYTOREMOVE: sometrash
replicas: '1'
version: 1.2.3
`
		pathParts := []string{"MYAPP_KEYTOREMOVE"}
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlRemoveEntryRecursive(&yamlContent, pathParts)
		assert.NotNil(t, err)
		assert.Contains(t, "No such path in target YAML document", err.Error())

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.Contains(t, changedyaml, "MYAPP_KEYTOREMOVE")
		assert.Contains(t, changedyaml, "sometrash")
	})
	t.Run("Should remove subtree if it matches the key", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
cluster: utv
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: '1'
version: 1.2.3
`
		pathParts := []string{"config"}
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlRemoveEntryRecursive(&yamlContent, pathParts)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.NotContains(t, changedyaml, "config")
		assert.NotContains(t, changedyaml, "MYAPP_SOME_KEY")
		assert.NotContains(t, changedyaml, "somevalue")
		assert.NotContains(t, changedyaml, "MYAPP_SOME_OTHER_KEY")
		assert.NotContains(t, changedyaml, "someothervalue")
		assert.Equal(t, 4, len(yamlContent))
	})
}

func Test_yamlSetOrCreateRecursive_Do(t *testing.T) {
	t.Run("Should set value on normal yaml content", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
cluster: utv
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: '1'
version: 1.2.3
`
		pathParts := []string{"config", "MYAPP_NEW_KEY"}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.Contains(t, changedyaml, "MYAPP_NEW_KEY")
		assert.Contains(t, changedyaml, "newValue")
	})

	t.Run("Should set value on minimal yaml content", func(t *testing.T) {
		content := `---`
		expected := "MYAPP_NEW_KEY: newValue\n"
		pathParts := []string{"MYAPP_NEW_KEY"}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)
		if len(yamlContent) == 0 {
			yamlContent = make(map[interface{}]interface{})
		}

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.Contains(t, changedyaml, "MYAPP_NEW_KEY")
		assert.Contains(t, changedyaml, "newValue")
		assert.Equal(t, expected, changedyaml)
	})

	t.Run("Should set new value on new multi level path", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
`
		expected := "baseFile: myapp.yaml\nfirst:\n  second:\n    MYAPP_NEW_KEY: newValue\n"
		pathParts := []string{"first", "second", "MYAPP_NEW_KEY"}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.Equal(t, expected, changedyaml)
	})

	t.Run("Should replace existing value", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
config:
  MYAPP_SOME_KEY: somevalue
`
		expected := "baseFile: myapp.yaml\nconfig:\n  MYAPP_SOME_KEY: newValue\n"
		pathParts := []string{"config", "MYAPP_SOME_KEY"}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.Nil(t, err)

		changedyamlbytearray, err := yaml.Marshal(yamlContent)
		assert.Nil(t, err)
		changedyaml := string(changedyamlbytearray)
		assert.NotNil(t, changedyaml)
		assert.Contains(t, changedyaml, "newValue")
		assert.NotContains(t, changedyaml, "somevalue")
		assert.Equal(t, expected, changedyaml)
	})

	t.Run("Should fail with empty path", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
config:
  MYAPP_SOME_KEY: somevalue
`
		pathParts := []string{}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.NotNil(t, err)
		assert.Equal(t, "Path can not be empty", err.Error())
	})

	t.Run("Should fail on numeric entry in path", func(t *testing.T) {
		content := `---
baseFile: myapp.yaml
config:
  MYAPP_SOME_KEY: somevalue
`
		pathParts := []string{"config", "270"}
		value := "newValue"
		var yamlContent map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(content), &yamlContent)
		assert.Nil(t, err)

		err = yamlSetOrCreateRecursive(&yamlContent, pathParts, value)
		assert.NotNil(t, err)
		assert.Equal(t, "Path can not have numeric entries", err.Error())
	})
}
