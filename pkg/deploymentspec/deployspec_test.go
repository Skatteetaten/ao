package deploymentspec

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readTestFile(t *testing.T) *DeploymentSpec {
	filePath := "./test_files/deployspec.json"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	var deploySpec DeploymentSpec
	err = json.Unmarshal(data, &deploySpec)
	if err != nil {
		t.Fatal(err)
	}

	return &deploySpec
}

func Test_GetString(t *testing.T) {
	deploySpec := readTestFile(t)

	assert.Equal(t, "east", deploySpec.GetString("cluster"))
	assert.Equal(t, "east", deploySpec.GetString("/cluster"))
	assert.Equal(t, "actuator", deploySpec.GetString("/management/path"))
	assert.Equal(t, "200m", deploySpec.GetString("/resources/cpu/max"))
}

func Test_GetStringFromNull(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, "-", deploySpec.GetString("/does/not/exist"))
}

func Test_GetStringFromNumber(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, "8080", deploySpec.GetString("/readiness/port"))
}

func Test_GetStringFromBool(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, "true", deploySpec.GetString("/prometheus"))
}

func Test_GetBool(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, true, deploySpec.GetBool("/prometheus"))
}

func Test_GetBoolFromNull(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, false, deploySpec.GetBool("/does/not/exist"))
}

func Test_CustomFunctions(t *testing.T) {
	deploySpec := readTestFile(t)
	assert.Equal(t, "east", deploySpec.Cluster())
	assert.Equal(t, "dev", deploySpec.Environment())
	assert.Equal(t, "flubber", deploySpec.Name())
	assert.Equal(t, "1", deploySpec.Version())
}

func Test_NonExistingField(t *testing.T) {
	assert.Equal(t, "-", readTestFile(t).GetString("doesnotexist"))
}

func Test_NewDeploymentSpec(t *testing.T) {
	deploySpec := NewDeploymentSpec("flubber", "dev", "east", "1")
	assert.Equal(t, "flubber", deploySpec.Name())
	assert.Equal(t, "dev", deploySpec.Environment())
	assert.Equal(t, "east", deploySpec.Cluster())
	assert.Equal(t, "1", deploySpec.Version())
	assert.Equal(t, "-", deploySpec.GetString("/does/not/exist"))
}
