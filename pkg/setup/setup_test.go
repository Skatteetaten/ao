package setup

import (
	"testing"
	"io/ioutil"
	"strings"
	"unicode"
)

func TestExecuteSetup(t *testing.T) {
	//var expected int = OPERATION_OKs
	var args []string = make([]string, 1)
	var overrideFiles []string
	var err error
	var out string

	args[0] = "testfiles/utv/about.json"
	out, err = ExecuteSetup(args, true, false, false, false, false, overrideFiles)
	if err != nil {
		t.Errorf("Failed simple dryrun: %v", err.Error())
	} else {
		fileJson, _ := ioutil.ReadFile("testresult.json")
		if strings.Compare(stripSpaces(out), stripSpaces(string(fileJson))) != 0 {
			t.Errorf("Dryrun result different than expected")
		}
	}

	args = make([]string, 2)
	args[0] = "testfiles/utv/about.json"
	args[1] = "{\"Game\": \"Thrones\"}"
	_, err = ExecuteSetup(args, true, false, false, false, false, overrideFiles)
	if err == nil {
		t.Errorf("Did not detect missing file reference")
	}

	overrideFiles = make([]string, 2)
	overrideFiles[0] = "File1"
	overrideFiles[1] = "File2"

	_, err = ExecuteSetup(args, true, false, false, false, false, overrideFiles)
	if err == nil {
		t.Errorf("Did not detect missing configuration")
	}
}

func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}
