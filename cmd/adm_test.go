package cmd

import (
	"bytes"
	"github.com/skatteetaten/ao/pkg/session"
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

	AOConfig = GetDefaultAOConfig()
	AOSession = &session.AOSession{
		RefName:      "",
		APICluster:   "utv01",
		AuroraConfig: "",
		Localhost:    false,
		Tokens:       map[string]string{},
	}

	for _, tc := range cases {
		buffer := &bytes.Buffer{}
		testCommand.SetOutput(buffer)

		flagShowAll = tc.FlagShowAll
		printClusters(testCommand, []string{})

		fileName := "test_files/" + tc.ExpectedStringFile

		if *updateFiles {
			ioutil.WriteFile(fileName, buffer.Bytes(), 644)
		}

		data, _ := ioutil.ReadFile(fileName)
		assert.Equal(t, string(data), buffer.String())
	}
}
