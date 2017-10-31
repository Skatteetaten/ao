package fuzzy

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

var fileNames = []string{
	"about.json",
	"console.json",
	"boober.json",
	"utv/about.json",
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

func TestFindMatches(t *testing.T) {

	tests := []struct {
		Search     string
		WithSuffix bool
		Expected   []string
	}{
		{"about", true, []string{"about.json"}},
		{"console", true, []string{"console.json"}},
		{"test/boober", true, []string{"test/boober.json"}},
		{"con", true, []string{"console.json", "utv/console.json", "test/console.json"}},
		{"con", false, []string{"console", "utv/console", "test/console"}},
		{"utv/ab", true, []string{"utv/about.json", "utv-relay/about.json"}},
		{"utv/o", true, []string{"utv/about.json", "utv/boober.json", "utv/console.json",
			"utv-relay/about.json", "utv-relay/boober.json"}},
	}

	for _, test := range tests {
		matches, err := FindMatches(test.Search, fileNames, test.WithSuffix)
		assert.Equal(t, matches, test.Expected, test.Search+" returned more matches than expected.")

		if err != nil {
			t.Error(err)
		}

		for i, m := range matches {
			assert.Equal(t, m, test.Expected[i])
		}
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

		assert.Equal(t, filename, test.Expected)
	}
}

func TestFindAllFor(t *testing.T) {
	tests := []struct {
		Search   string
		Mode     DeployFilterMode
		Expected []string
	}{
		{"utv", ENV, []string{"boober", "console"}},
		{"console", APP, []string{"test", "utv"}},
		{"test", ENV, []string{"boober", "console"}},
		{"test-r", ENV, []string{}},
		{"boober", APP, []string{"test", "test-relay", "utv", "utv-relay"}},
		{"boo", APP, []string{}},
	}

	filteredFiles := FilterFileNamesForDeploy(fileNames)

	for _, test := range tests {
		deploys, err := FindAllDeploysFor(test.Mode, test.Search, filteredFiles)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, deploys, test.Expected)
	}
}
