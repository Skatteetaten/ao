package fuzzy

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var fileNames = []string{
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
}

func TestFilterFileNamesForDeploy(t *testing.T) {
	var expected = []string{
		"utv/boober",
		"utv/console",
		"utv-relay/boober",
		"test/boober",
		"test/console",
		"test-relay/boober",
	}

	actual := FilterFileNamesForDeploy(fileNames)
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
	}

	for _, test := range tests {
		matches, _ := FindMatches(test.Search, fileNames, test.WithSuffix)
		assert.Equal(t, test.Expected, matches, test.Search+" returned unexpected matches than expected.")
	}
}

func TestFindFileToEdit(t *testing.T) {
	tests := []struct {
		Search   string
		Prompt   bool
		Expected string
	}{
		{"about", false, "about.json"},
		{"console", false, "console.json"},
		{"utv/ab", false, "utv/about.json"},
	}

	for _, test := range tests {
		filename, err := FindFileToEdit(test.Search, fileNames, test.Prompt)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, test.Expected, filename)
	}
}

func TestFindApplicationsToDeploy(t *testing.T) {
	tests := []struct {
		Search   string
		Expected []string
	}{
		{"aosdfkja",  []string{}},
		{"utv",  []string{"utv/boober", "utv/console"}},
		{"console",  []string{"test/console", "utv/console"}},
		{"test",  []string{"test/boober", "test/console"}},
		{"test-r",  []string{"test-relay/boober"}},
		{"boober",  []string{"test/boober", "test-relay/boober", "utv/boober", "utv-relay/boober"}},
		{"boo",  []string{"utv/boober", "test/boober", "utv-relay/boober", "test-relay/boober"}},
	}

	filteredFiles := FilterFileNamesForDeploy(fileNames)

	for _, test := range tests {
		deploys, _ := FindApplicationsToDeploy(test.Search, filteredFiles, false)
		assert.Equal(t, test.Expected, deploys, "Searching for " + test.Search)
	}
}

func TestFindAllFor(t *testing.T) {
	tests := []struct {
		Search   string
		Mode     DeployFilterMode
		Expected []string
	}{
		{"utv", ENV_FILTER, []string{"utv/boober", "utv/console"}},
		{"console", APP_FILTER, []string{"test/console", "utv/console"}},
		{"test", ENV_FILTER, []string{"test/boober", "test/console"}},
		{"test-r", ENV_FILTER, []string{}},
		{"boober", APP_FILTER, []string{"test/boober", "test-relay/boober", "utv/boober", "utv-relay/boober"}},
		{"boo", APP_FILTER, []string{}},
	}

	filteredFiles := FilterFileNamesForDeploy(fileNames)

	for _, test := range tests {
		deploys := FindAllDeploysFor(test.Mode, test.Search, filteredFiles)
		assert.Equal(t, test.Expected, deploys)
	}
}
