package boober

import (
	"encoding/json"
	"reflect"
	"runtime"
	"testing"
)

func TestCallBooberInstance(t *testing.T) {
	var expected int = booberError
	var res int
	const illegalUrl string = "https://westeros.skatteetaten.no/boober"

	res = CallBooberInstance("{\"Game\": \"Thrones\"}", false,
		illegalUrl)
	if res != expected {
		t.Error("Did not detect illegal URL")
	}
}

func TestFolder2Map(t *testing.T) {
	var expected int = 2
	var res map[string]json.RawMessage
	res = Folder2Map("testfiles/utv", "")
	if len(res) != expected {
		t.Errorf("Returned map with length %v, expected %v", len(res), expected)
	}
	if res["about.json"] == nil {
		t.Error("Did not map about.json file")
	}
}

func TestExecuteSetup(t *testing.T) {
	//var expected int = OPERATION_OKs
	var args []string = make([]string, 1)
	var overrideFiles []string
	var res, expected int

	args[0] = "testfiles/utv/about.json"
	expected = operationOk
	res = ExecuteSetup(args, true, false, overrideFiles)
	if res != expected {
		t.Errorf("Failed simple dryrun: %v", res)
	}

	args = make([]string, 2)
	args[0] = "testfiles/utv/about.json"
	args[1] = "{\"Game\": \"Thrones\"}"
	expected = MissingFileReference
	res = ExecuteSetup(args, true, false, overrideFiles)
	if res != expected {
		t.Errorf("Did not detect missing file reference: %v", res)
	}

	overrideFiles = make([]string, 2)
	overrideFiles[0] = "File1"
	overrideFiles[1] = "File2"
	expected = MissingConfiguration
	res = ExecuteSetup(args, true, false, overrideFiles)
	if res != expected {
		t.Errorf("Did not detect missing configuration: %v", res)
	}
}

func TestCombineMaps(t *testing.T) {
	var expected map[string]json.RawMessage = make(map[string]json.RawMessage)
	var map1 map[string]json.RawMessage = make(map[string]json.RawMessage)
	var map2 map[string]json.RawMessage = make(map[string]json.RawMessage)
	var res map[string]json.RawMessage

	map1["File1"] = json.RawMessage("{\"Game\": \"Thrones\"}")
	map2["File2"] = json.RawMessage("{\"Kingslayer\": \"Jamie Lannister\"}")

	expected["File1"] = json.RawMessage("{\"Game\": \"Thrones\"}")
	expected["File2"] = json.RawMessage("{\"Kingslayer\": \"Jamie Lannister\"}")

	res = CombineMaps(map1, map2)
	if !reflect.DeepEqual(res, expected) {
		t.Error("Failed to combine maps")
	}
}

func TestIsLegalFileFolder(t *testing.T) {
	var expected, res int
	var testfolder string

	expected = specIsFolder
	if runtime.GOOS == "windows" {
		testfolder = "C:\\Windows"
	} else {
		testfolder = "/bin"
	}
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize bin/windows as folder")
	}

	expected = specIsFile
	if runtime.GOOS == "windows" {
		testfolder = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	} else {
		testfolder = "/etc/hosts"
	}
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize hosts file as legal file")
	}

	expected = SpecIllegal
	testfolder = "/Go/is/an/open/source/programming/language/that/makes/it/easy/to/build/" +
		"simple/reliable/and/efficient/software."
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize illegal folder")
	}
}

func TestLegalJson(t *testing.T) {
	var legalJsonString string
	var expected bool
	var res bool

	legalJsonString = `
	{
		"build": {
			"VERSION": "1.0.6-SNAPSHOT"
		},
		"deploy": {
			"DATABASE": "demo:5bfe8be8-cc73-4882-ab05-212ddbd10632"
		},
		"config": {
			"DEMO_PROPERTY": "ELVIS LIVES!"
		}
	}`

	expected = true
	res = IsLegalJson(legalJsonString)
	if res != expected {
		t.Error("Did not recognize legal JSON")
	}

	illegalJsonString := `
	{
		"build":
			"VERSION": "1.0.6-SNAPSHOT"
		},
		"deploy": {
			"DATABASE": "demo:5bfe8be8-cc73-4882-ab05-212ddbd10632"
		},
		"config": {
			"DEMO_PROPERTY": "ELVIS LIVES!"
		}
	}`

	expected = false
	res = IsLegalJson(illegalJsonString)
	if res != expected {
		t.Error("Did not recognize illegal JSON")
	}

	illegalJsonString = `
	{
		"build": {
			"VERSION": "1.0.6-SNAPSHOT"
		},
		"deploy": {
			"DATABASE": "demo:5bfe8be8-cc73-4882-ab05-212ddbd10632"
		},
		"config": {
			"DEMO_PROPERTY" "ELVIS LIVES!"
		}
	}`

	expected = false
	res = IsLegalJson(illegalJsonString)
	if res != expected {
		t.Error("Did not recognize illegal JSON")
	}

	illegalJsonString = `
	<build VERSION=1.0.6-SNAPSHOT>
		<deploy>
			<DATABASE>demo:5bfe8be8-cc73-4882-ab05-212ddbd10632</DATABASE>
		</deploy>
		<config DEMO_PROPERTY="ELVIS LIVES!">
		</deploy>
	</build>`

	expected = false
	res = IsLegalJson(illegalJsonString)
	if res != expected {
		t.Error("Did not recognize XML as illegal JSON")
	}
}

func TestPrettyPrintJson(t *testing.T) {
	legalJsonString := `{"build": {"VERSION": "1.0.6-SNAPSHOT"},"deploy": {
	"DATABASE": "demo:5bfe8be8-cc73-4882-ab05-212ddbd10632"},"config": {
	"DEMO_PROPERTY": "ELVIS LIVES!"}}`
	expected :=
		`{
	"build": {
		"VERSION": "1.0.6-SNAPSHOT"
	},
	"deploy": {
		"DATABASE": "demo:5bfe8be8-cc73-4882-ab05-212ddbd10632"
	},
	"config": {
		"DEMO_PROPERTY": "ELVIS LIVES!"
	}
}`

	res := PrettyPrintJson(legalJsonString)
	if res != expected {
		t.Error("Did not pretty print correctly")
	}
}
