package auroraconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var fileNames = FileNames{
	"about.json",
	"console.json",
	"boober.json",
	"utv/about.json",
	"utv/about-template.json",
	"utv/boober.json",
	"utv/console.json",
	"utv-relay/about.json",
	"utv-relay/boober.json",
	"test/about.json",
	"test/boober.json",
	"test/console.json",
	"test-relay/about.json",
	"test-relay/boober.json",
	"test-relay/app2.yaml",
}

func TestFilterFileNamesForDeploy(t *testing.T) {
	var expected = []string{
		"test-relay/app2",
		"test-relay/boober",
		"test/boober",
		"test/console",
		"utv-relay/boober",
		"utv/boober",
		"utv/console",
	}

	actual := fileNames.GetApplicationDeploymentRefs()
	assert.Equal(t, expected, actual)
}

func TestFindMatches(t *testing.T) {

	tests := []struct {
		Search     string
		WithSuffix bool
		Expected   []string
	}{
		{"asdlfkja", true, []string{}},
		{"about", true, []string{"about.json"}},
		{"console", true, []string{"console.json"}},
		{"test/boober", true, []string{"test/boober.json"}},
		{"con", true, []string{"console.json", "utv/console.json", "test/console.json"}},
		{"con", false, []string{"console", "utv/console", "test/console"}},
		{"utv/ab", true, []string{"utv/about.json", "utv-relay/about.json", "utv/about-template.json"}},
		{"utv/o", true, []string{"utv/about.json", "utv/boober.json", "utv/console.json",
			"utv-relay/about.json", "utv-relay/boober.json", "utv/about-template.json"}},
		{"testboober", false, []string{}},
		{"test/boo", false, []string{"test/boober", "test-relay/boober"}},
		{"test-relay2", false, []string{}},
		{"test-relay2/app2", false, []string{}},
		{"test-relay/app2", false, []string{"test-relay/app2"}},
	}

	for _, test := range tests {
		matches := FindMatches(test.Search, fileNames, test.WithSuffix)
		assert.Equal(t, test.Expected, matches, test.Search+" returned other matches than expected.")
	}
}

func TestFindFileToEdit(t *testing.T) {
	tests := []struct {
		Search   string
		Expected []string
	}{
		{"about", []string{"about.json"}},
		{"console", []string{"console.json"}},
		{"utv/ab", []string{"utv/about.json", "utv-relay/about.json", "utv/about-template.json"}},
		{"test-relay2", []string{}},
	}

	for _, test := range tests {
		filename := SearchForFile(test.Search, fileNames)
		assert.Equal(t, test.Expected, filename)
	}
}

func TestFindApplicationsToDeploy(t *testing.T) {
	tests := []struct {
		Search   string
		Expected []string
	}{
		{"aosdfkja", []string{}},
		{"utv", []string{"utv/boober", "utv/console"}},
		{"console", []string{"test/console", "utv/console"}},
		{"test", []string{"test/boober", "test/console"}},
		{"test-r", []string{"test-relay/app2", "test-relay/boober"}},
		{"boober", []string{"test/boober", "test-relay/boober", "utv/boober", "utv-relay/boober"}},
		{"boo", []string{"utv/boober", "test/boober", "utv-relay/boober", "test-relay/boober"}},
		{"test-relay2", []string{}},
	}

	filteredFiles := fileNames.GetApplicationDeploymentRefs()

	for _, test := range tests {
		deploys := SearchForApplications(test.Search, filteredFiles)
		assert.Equal(t, test.Expected, deploys, "Searching for "+test.Search)
	}
}

func TestFindAllFor(t *testing.T) {
	tests := []struct {
		Search   string
		Mode     FilterMode
		Expected []string
	}{
		{"utv", EnvFilter, []string{"utv/boober", "utv/console"}},
		{"console", AppFilter, []string{"test/console", "utv/console"}},
		{"test", EnvFilter, []string{"test/boober", "test/console"}},
		{"test-r", EnvFilter, []string{}},
		{"boober", AppFilter, []string{"test/boober", "test-relay/boober", "utv/boober", "utv-relay/boober"}},
		{"boo", AppFilter, []string{}},
		{"test-relay2", AppFilter, []string{}},
	}

	filteredFiles := fileNames.GetApplicationDeploymentRefs()

	filteredFiles = append(filteredFiles, "illegalfile")

	for _, test := range tests {
		deploys := FindAllDeploysFor(test.Mode, test.Search, filteredFiles)
		assert.Equal(t, test.Expected, deploys)
	}
}
