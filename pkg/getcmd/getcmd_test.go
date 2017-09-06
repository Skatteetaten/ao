package getcmd

import (
	"encoding/json"
	"testing"

	"github.com/skatteetaten/ao/pkg/serverapi"
)

func TestFormatFileList(t *testing.T) {
	var files []string
	files = make([]string, 2)
	files[0] = "fil1"
	files[1] = "fil2"

	expected := "NAME\nfil1\nfil2"
	result := formatFileList(files)

	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}

}

func TestGetFileList(t *testing.T) {
	var auroraConfig serverapi.AuroraConfig
	auroraConfig.Files = make(map[string]json.RawMessage)
	auroraConfig.Files["fil1.json"] = json.RawMessage("{\"foo\":\"bar\"}")
	auroraConfig.Files["fil2.json"] = json.RawMessage("{\"foo\":\"bar\"}")

	result := getFileList(&auroraConfig)
	if len(result) != len(auroraConfig.Files) {
		t.Errorf("Expected length %v, got %v", len(result), len(auroraConfig.Files))
	}

	for i := range result {
		_, exists := auroraConfig.Files[result[i]]
		if !exists {
			t.Errorf("File %v does not exist in auroraConfig", result[i])
		}
	}
}
