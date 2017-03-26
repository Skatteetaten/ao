package jsonutil

import (
	"testing"
	"encoding/json"
	"reflect"
)

func TestFolder2Map(t *testing.T) {
	var expected int = 2
	var res map[string]json.RawMessage
	res, err := Folder2Map("testfiles/utv", "")
	if err != nil {
		t.Errorf("Folder2Map returned an error: %v", err.Error())
	} else {
		if len(res) != expected {
			t.Errorf("Returned map with length %v, expected %v", len(res), expected)
		}
		if res["about.json"] == nil {
			t.Error("Did not map about.json file")
		}
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
