package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getValidFileNameFromPath(t *testing.T) {
	cases := []struct {
		Path     string
		Expected string
	}{
		{"about.json", "about.json"},
		{"./about.json", "about.json"},
		{"~/about.json", "about.json"},
		{"test/foo.json", "test/foo.json"},
		{"./test/foo.json", "test/foo.json"},
		{"home/projects/prod/about.json", "prod/about.json"},
	}

	for _, tc := range cases {
		actual := getValidFileNameFromPath(tc.Path)
		assert.Equal(t, tc.Expected, actual)
	}
}
