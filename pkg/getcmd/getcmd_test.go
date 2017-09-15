package getcmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/skatteetaten/ao/pkg/serverapi"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/jsonutil"
)

func TestDeployments(t *testing.T) {
	//var getcmd *GetcmdClass
	//getcmd = new(GetcmdClass)

	// TODO: Refactor method
}

func TestFormatDeploymentList(t *testing.T) {
	var envList []string
	var appList []fuzzyargs.LegalDeployStruct

	envList = make([]string, 2)
	appList = make([]fuzzyargs.LegalDeployStruct, 2)

	envList[0] = "test"
	envList[1] = "prod"

	appList[0].AppName = "myapp"
	appList[0].EnvName = "test"
	appList[1].AppName = "myotherapp"
	appList[1].EnvName = "prod"

	deploymentList := formatDeploymentList("", envList, appList)
	if !strings.Contains(deploymentList, "test") {
		t.Errorf("Deploymentlist does not contain test env")
	}

	if !strings.Contains(deploymentList, "ENV") {
		t.Errorf("Deploymentlist missing header ENV")
	}
}

func TestApps (t *testing.T) {
	// TODO: Refactor method
}

func TestFormatAppList (t *testing.T) {
	var appList []string
	appList = make([]string, 1)
	appList[0] = "myapp"

	output := formatAppList(appList)
	if !strings.Contains(output, "myapp") {
		t.Errorf("Missing app: %v", "myapp")
	}
}


func TestEnvs(t *testing.T) {
	// TODO: Refactor method
}

func TestFormatEnvList (t *testing.T) {
	var envList []string
	envList = make([]string, 1)
	envList[0] = "myenv"

	output := formatEnvList(envList)
	if !strings.Contains(output, "myenv") {
		t.Errorf("Missing env: %v", "myenv")
	}
}

func TestFiles(t *testing.T) {
	// TODO; Refactor method
}

func TestFormatFileList(t *testing.T) {
	var files []string
	files = make([]string, 2)
	files[0] = "fil1"
	files[1] = "fil2"

	expected := jsonutil.StripSpaces("FILE/FOLDERFILE\nfil1\nfil2")
	result := jsonutil.StripSpaces(formatFileList(files))

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

func TestFile(t *testing.T) {
	// TODO: Refactor
}

func TestClusters(t *testing.T) {
	// TODO: Refactor
}

func TestVaults(t *testing.T) {
	// TODO: Refactor
}

func TestVault(t *testing.T) {
	// TODO: Refactor
}

func TestSecret(t *testing.T) {
	// TODO: Refactor
}

func TestKubeConfig(t *testing.T) {
	// TODO: Refactor
}

func TestOcLogin(t *testing.T) {
	// TODO: Refactor
}

