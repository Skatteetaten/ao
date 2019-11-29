package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintClusters(t *testing.T) {
	cases := []struct {
		FlagShowAll        bool
		ExpectedStringFile string
	}{
		{false, "adm_clusters.txt"},
		{true, "adm_clusters_all.txt"},
	}

	AO = GetDefaultAOConfig()

	for _, tc := range cases {
		buffer := &bytes.Buffer{}
		testCommand.SetOutput(buffer)

		flagShowAll = tc.FlagShowAll
		PrintClusters(testCommand, []string{})

		fileName := "test_files/" + tc.ExpectedStringFile

		if *updateFiles {
			ioutil.WriteFile(fileName, buffer.Bytes(), 644)
		}

		data, _ := ioutil.ReadFile(fileName)
		assert.Equal(t, string(data), buffer.String())
	}
}
