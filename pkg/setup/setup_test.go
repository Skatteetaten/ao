package setup

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"unicode"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
)

func TestExecuteSetup(t *testing.T) {
	//var expected int = OPERATION_OKs
	var args []string = make([]string, 1)
	var overrideFiles []string
	var err error
	var out string
	var persistentOptions cmdoptions.CommonCommandOptions

	args[0] = "testfiles/utv/about.json"
	persistentOptions.DryRun = true
	out, err = ExecuteSetup(args, overrideFiles, &persistentOptions)
	if err != nil {
		t.Errorf("Failed simple dryrun: %v", err.Error())
	} else {
		fileJson, err := ioutil.ReadFile("testresult.json")
		if err != nil {
			t.Errorf("Unable to read testresult.json")
		}
		fileJsonStr := stripSpaces(string(fileJson))
		fmt.Println("Testresult:")
		fmt.Println(fileJsonStr)
		out := stripSpaces(out)
		fmt.Println("Out:")
		fmt.Println(out)
		if strings.Compare(out, fileJsonStr) != 0 {
			t.Errorf("Dryrun result different than expected")
		}
	}

	args = make([]string, 2)
	args[0] = "testfiles/utv/about.json"
	args[1] = "{\"Game\": \"Thrones\"}"
	_, err = ExecuteSetup(args, overrideFiles, &persistentOptions)
	if err == nil {
		t.Errorf("Did not detect missing file reference")
	}

	overrideFiles = make([]string, 2)
	overrideFiles[0] = "File1"
	overrideFiles[1] = "File2"

	_, err = ExecuteSetup(args, overrideFiles, &persistentOptions)
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
