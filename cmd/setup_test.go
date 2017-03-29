package cmd

import (
	"bytes"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)

func TestSetup(t *testing.T) {
	cmd := exec.Command("../bin/amd64/aoc", "setup", "../pkg/setup/testfiles/utv/about.json", "-d")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		t.Error(err.Error())
	}
	outputStr := out.String()
	if len(outputStr) == 0 {
		t.Error("No output of dryrun")
	}
	//t.Error(outputStr)

	fileJson, err := ioutil.ReadFile("../pkg/setup/testresult.json")
	if err != nil {
		t.Errorf("Unable to read testresult.json")
	}
	testresultStr := jsonutil.StripSpaces(string(fileJson))

	outputStr = jsonutil.StripSpaces(outputStr)
	if !strings.Contains(outputStr, testresultStr) {
		t.Errorf("Output does not contain expected test result: \n%v\n%v", outputStr, testresultStr)
	}
}
