package cmd

import (
	"bytes"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var files = client.FileNames{
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
		UpdateTestFile(fileName, buffer.Bytes())
		data, _ := ioutil.ReadFile(fileName)

		assert.Equal(t, string(data), buffer.String())
	}
}

func TestGetApplicationIdTable(t *testing.T) {

	header, rows := GetApplicationIdTable(files.GetApplicationIds())

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
