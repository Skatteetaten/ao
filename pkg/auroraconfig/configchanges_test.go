package auroraconfig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SetValue_Do(t *testing.T) {
	t.Run("Should set value in Json AuroraConfigFile (happy test)", func(t *testing.T) {
		content := `{
		    "baseFile": "myapp.json"
		}`
		auroraConfigFile := File{
			Name:     "myconfigfile.json",
			Contents: content,
		}
		path := "/config/MYAPP_NEW_KEY"
		value := "newValue"

		err := SetValue(&auroraConfigFile, path, value)
		assert.Nil(t, err)

		changedjson := auroraConfigFile.Contents
		assert.NotNil(t, changedjson)
		assert.Contains(t, changedjson, "MYAPP_NEW_KEY")
		assert.Contains(t, changedjson, "newValue")
		assert.Equal(t, "{\n  \"baseFile\": \"myapp.json\",\n  \"config\": {\n    \"MYAPP_NEW_KEY\": \"newValue\"\n  }\n}\n", changedjson)
	})
	t.Run("Should set value in yaml AuroraConfigFile (happy test)", func(t *testing.T) {
		content := `---
baseFile: myapp.json
cluster: utv01
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: '1'
version: 1.2.3
`
		expected := `---
baseFile: myapp.json
cluster: utv01
config:
  MYAPP_NEW_KEY: newValue
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: "1"
version: 1.2.3
`
		auroraConfigFile := File{
			Name:     "myconfigfile.yaml",
			Contents: content,
		}
		path := "/config/MYAPP_NEW_KEY"
		value := "newValue"

		err := SetValue(&auroraConfigFile, path, value)
		assert.Nil(t, err)

		changedyaml := auroraConfigFile.Contents
		assert.NotNil(t, changedyaml)
		assert.Contains(t, changedyaml, "MYAPP_NEW_KEY")
		assert.Contains(t, changedyaml, "newValue")

		assert.Equal(t, expected, changedyaml)
	})
}

func Test_RemoveEntry_Do(t *testing.T) {
	t.Run("Should remove value from Json AuroraConfigFile", func(t *testing.T) {
		content := `{
            "baseFile": "myapp.json",
            "cluster": "utv01",
            "config": {
                "MYAPP_SOME_KEY": "somevalue",
                "MYAPP_SOME_OTHER_KEY": "someothervalue",
                "MYAPP_KEYTOREMOVE": "sometrash"
            },
            "replicas": "1",
            "version": "1.2.3"
        }`
		expected := `{
  "baseFile": "myapp.json",
  "cluster": "utv01",
  "config": {
    "MYAPP_SOME_KEY": "somevalue",
    "MYAPP_SOME_OTHER_KEY": "someothervalue"
  },
  "replicas": "1",
  "version": "1.2.3"
}
`
		auroraConfigFile := File{
			Name:     "myconfigfile.json",
			Contents: content,
		}
		path := "/config/MYAPP_KEYTOREMOVE"

		err := RemoveEntry(&auroraConfigFile, path)
		assert.Nil(t, err)

		changedjson := auroraConfigFile.Contents
		assert.NotNil(t, changedjson)
		assert.NotContains(t, changedjson, "MYAPP_SOMEKEYTOREMOVE")
		assert.NotContains(t, changedjson, "valuetoremove")
		assert.Equal(t, expected, changedjson)
	})
	t.Run("Should remove value from Yaml AuroraConfigFile", func(t *testing.T) {
		content := `---
baseFile: myapp.json
cluster: utv01
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
  MYAPP_SOMEKEYTOREMOVE: valuetoremove
replicas: '1'
version: 1.2.3
`
		expected := `---
baseFile: myapp.json
cluster: utv01
config:
  MYAPP_SOME_KEY: somevalue
  MYAPP_SOME_OTHER_KEY: someothervalue
replicas: "1"
version: 1.2.3
`
		auroraConfigFile := File{
			Name:     "myconfigfile.yml",
			Contents: content,
		}
		path := "/config/MYAPP_SOMEKEYTOREMOVE"

		err := RemoveEntry(&auroraConfigFile, path)
		assert.Nil(t, err)

		changedyaml := auroraConfigFile.Contents
		assert.NotNil(t, changedyaml)
		assert.NotContains(t, changedyaml, "MYAPP_SOMEKEYTOREMOVE")
		assert.NotContains(t, changedyaml, "valuetoremove")
		assert.Equal(t, expected, changedyaml)
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
