package setup

import (
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"io/ioutil"
	"strings"
	"testing"
)

func TestExecuteSetup(t *testing.T) {
	//var expected int = OPERATION_OKs
	var args []string = make([]string, 1)
	var overrideFiles []string
	var err error
	var out string
	var persistentOptions cmdoptions.CommonCommandOptions
	var setupObject SetupClass

	args[0] = "testfiles/utv/about.json"
	persistentOptions.DryRun = true
	out, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions)
	if err != nil {
		t.Errorf("Failed simple dryrun: %v", err.Error())
	} else {
		fileJson, err := ioutil.ReadFile("testresult.json")
		if err != nil {
			t.Errorf("Unable to read testresult.json")
		}
		fileJsonStr := jsonutil.StripSpaces(string(fileJson))
		fmt.Println("Testresult:")
		fmt.Println(fileJsonStr)
		out := jsonutil.StripSpaces(out)
		fmt.Println("Out:")
		fmt.Println(out)
		if strings.Compare(out, fileJsonStr) != 0 {
			t.Errorf("Dryrun result different than expected")
		}
	}

	args = make([]string, 2)
	args[0] = "testfiles/utv/about.json"
	args[1] = "{\"Game\": \"Thrones\"}"
	_, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions)
	if err == nil {
		t.Errorf("Did not detect missing file reference")
	}

	overrideFiles = make([]string, 2)
	overrideFiles[0] = "File1"
	overrideFiles[1] = "File2"

	_, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions)
	if err == nil {
		t.Errorf("Did not detect missing configuration")
	}
}
