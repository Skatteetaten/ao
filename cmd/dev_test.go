package cmd

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/skatteetaten/ao/pkg/session"
	architect "github.com/skatteetaten/architect/v2/pkg/build"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"testing"
)

type DevMock struct {
	mock.Mock
	deployType              string
	numberOfFileNameMatches int
}

func (api *DevMock) GetFileNames() (auroraconfig.FileNames, error) {
	var filenames = make(auroraconfig.FileNames, api.numberOfFileNameMatches)
	for i := 0; i < api.numberOfFileNameMatches; i++ {
		filenames[i] = "dev-utv/whoami"
	}
	return filenames, nil
}

func (api *DevMock) GetAuroraDeploySpec(applications []string, defaults bool, ignoreErrors bool) ([]deploymentspec.DeploymentSpec, error) {
	file := ReadTestFile("deployspec_response")
	var deployspec deploymentspec.DeploymentSpec
	err := json.Unmarshal(file, &deployspec)
	if err != nil {
		log.Fatal("Error with unmarshal")
	}

	if api.deployType != "development" {
		deployspec["type"] = map[string]interface{}{"value": "deploy"}
	}

	return []deploymentspec.DeploymentSpec{deployspec}, nil
}

func ReadTestFile(name string) []byte {
	filePath := fmt.Sprintf("./test_files/%s.json", name)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return data
}
func Test_Rollout(t *testing.T) {
	flagDevNoPrompt = true
	defer func() { flagDevNoPrompt = false }()

	devMock := &DevMock{deployType: "development", numberOfFileNameMatches: 1}
	AOSession = &session.AOSession{
		RefName:      "master",
		APICluster:   "testApiCluster",
		AuroraConfig: "",
		Localhost:    false,
		Tokens:       map[string]string{"east": "tokenhere"},
	}

	t.Run("Should build image", func(t *testing.T) {
		err := executeTest(devMock)

		assert.NoError(t, err)
	})

	t.Run("development type should be required", func(t *testing.T) {
		devMock.deployType = "deploy"

		err := executeTest(devMock)

		assert.ErrorContains(t, err, "You need to specify type=development")
	})
	t.Run("file should exist", func(t *testing.T) {
		err := executeTest(devMock, "whoami", "nonexistingfile.txt")

		assert.ErrorContains(t, err, "the provided path and name of the leveransepakke file does not exist")
	})

	t.Run("Should return error when more than one match", func(t *testing.T) {
		devMock.numberOfFileNameMatches = 2

		err := executeTest(devMock)
		assert.ErrorContains(t, err, "Search must be more specific")

	})
	t.Run("Should return error when no matches", func(t *testing.T) {
		devMock.numberOfFileNameMatches = 0

		err := executeTest(devMock)

		assert.ErrorContains(t, err, "No matches for whoami")

	})
}

func executeTest(devMock *DevMock, args ...string) error {
	testCmd := cobra.Command{RunE: func(cmd *cobra.Command, args []string) error {
		return Rollout(cmd, args, func(c architect.Configuration) {
			return
		}, devMock)
	}}

	if len(args) == 0 {
		args = []string{"whoami", "dev.go"}
	}
	testCmd.SetArgs(args)

	return testCmd.Execute()
}
