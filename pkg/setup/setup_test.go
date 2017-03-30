package setup

import (
	"encoding/json"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExecuteSetup(t *testing.T) {
	testLocalDryRun(t)
	//testLocalhostRun(t)
}

func testLocalDryRun(t *testing.T) {
	//var expected int = OPERATION_OKs
	var args []string = make([]string, 1)
	var overrideFiles []string
	var err error
	var out string
	var persistentOptions cmdoptions.CommonCommandOptions
	var setupObject SetupClass

	args[0] = "testfiles/utv/about.json"
	persistentOptions.DryRun = true
	out, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions, true)
	if err != nil {
		t.Errorf("Failed simple dryrun: %v", err.Error())
	} else {
		fileJson, err := ioutil.ReadFile("testresult.json")
		if err != nil {
			t.Errorf("Unable to read testresult.json")
		}
		fileJsonStr := jsonutil.StripSpaces(string(fileJson))

		out := jsonutil.StripSpaces(out)
		if strings.Compare(out, fileJsonStr) != 0 {
			t.Errorf("Dryrun result different than expected")
		}
	}

	args = make([]string, 2)
	args[0] = "testfiles/utv/about.json"
	args[1] = "{\"Game\": \"Thrones\"}"
	_, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions, true)
	if err == nil {
		t.Errorf("Did not detect missing file reference")
	}

	overrideFiles = make([]string, 2)
	overrideFiles[0] = "File1"
	overrideFiles[1] = "File2"

	_, err = setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions, true)
	if err == nil {
		t.Errorf("Did not detect missing configuration")
	}

}

var ch = make(chan jsonutil.ApiInferface)

func testLocalhostRun(t *testing.T) {
	//var result jsonutil.ApiInferface

	go listenToOneRequest(ch)

	time.Sleep(2)
	// Send request to localhost
	sendOneRequestToLocalhost(t)

	// Wait for result
	result := <-ch
	t.Error(result)
}

func sendOneRequestToLocalhost(t *testing.T) {
	var args []string = make([]string, 1)
	var overrideFiles []string
	var persistentOptions cmdoptions.CommonCommandOptions
	var setupObject SetupClass

	args[0] = "testfiles/utv/about.json"
	persistentOptions.Localhost = true
	out, err := setupObject.ExecuteSetup(args, overrideFiles, &persistentOptions, false)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Error(out)
	}
}

func listenToOneRequest(ch chan jsonutil.ApiInferface) {

	http.HandleFunc("/setup", handler)
	http.ListenAndServe("8080", nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
	var apiInterface jsonutil.ApiInferface

	apiInterface.Affiliation = "foobar"
	apiInterface.Env = "myenv"
	apiInterface.App = "myapp"
	js, err := json.Marshal(apiInterface)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	ch <- apiInterface
}
