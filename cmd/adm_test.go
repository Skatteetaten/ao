package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestPrintClusters(t *testing.T) {
	cases := []struct {
		FlagShowAll        bool
		ExpectedStringFile string
	}{
		{false, "adm_clusters.txt"},
		{true, "adm_clusters_all.txt"},
	}

	for _, tc := range cases {
		buffer := &bytes.Buffer{}
		testCommand.SetOutput(buffer)

		flagShowAll = tc.FlagShowAll
		PrintClusters(testCommand, []string{})

		fileName := "test_files/" + tc.ExpectedStringFile

		// Will update test file if update.files flag is set during testing
		UpdateTestFile(fileName, buffer.Bytes())

		data, _ := ioutil.ReadFile(fileName)
		assert.Equal(t, string(data), buffer.String())
	}
}
