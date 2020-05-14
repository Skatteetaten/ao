package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/stretchr/testify/assert"
)

var files = auroraconfig.FileNames{
	"bar.json",
	"baz.json",
	"foo/bar.json",
	"foo/baz.json",
	"about.json",
	"foo/about.json",
}

func TestDefaultTablePrinter(t *testing.T) {

	cases := []struct {
		Header   string
		Rows     []string
		NoHeader bool
		Expected string
	}{
		{"FILES", files, false, "table_with_header.txt"},
		{"FILES", files, true, "table_without_header.txt"},
	}

	for _, tc := range cases {
		// Persistent flag
		pFlagNoHeader = tc.NoHeader

		buffer := &bytes.Buffer{}
		DefaultTablePrinter(tc.Header, tc.Rows, buffer)

		fileName := "test_files/" + tc.Expected

		if *updateFiles {
			ioutil.WriteFile(fileName, buffer.Bytes(), 644)
		}

		data, _ := ioutil.ReadFile(fileName)
		assert.Equal(t, string(data), buffer.String())
	}
}

func TestGetApplicationDeploymentRefTable(t *testing.T) {

	header, rows := GetApplicationDeploymentRefTable(files.GetApplicationDeploymentRefs())

	assert.Equal(t, header, "ENVIRONMENT\tAPPLICATION")
	assert.Len(t, rows, 2)
	assert.Equal(t, rows[0], "foo\tbar")
	assert.Equal(t, rows[1], " \tbaz")
}

func TestGetFileRows(t *testing.T) {

	header, rows := GetFilesTable(files)

	assert.Equal(t, header, "FILES")
	assert.Len(t, rows, 6)
	assert.Equal(t, rows[0], "about.json")
	assert.Equal(t, rows[1], "bar.json")
	assert.Equal(t, rows[2], "baz.json")
	assert.Equal(t, rows[3], "foo/about.json")
	assert.Equal(t, rows[4], "foo/bar.json")
	assert.Equal(t, rows[5], "foo/baz.json")
}
