package boober

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const specIllegal = -1
const specIsFile = 1
const specIsFolder = 2

const booberNotInstalledResponse = "Application is not available"

// Struct to represent data to the Boober interface
type BooberInferface struct {
	Env         string                     `json:"env"`
	App         string                     `json:"app"`
	Affiliation string                     `json:"affiliation"`
	Files       map[string]json.RawMessage `json:"files"`
	Overrides   map[string]json.RawMessage `json:"overrides"`
}

// Structs to represent return data from the Boober interface
type BooberReturnObjects struct {
	Sources          json.RawMessage            `json:"sources"`
	Errors           []string                   `json:"errors"`
	Valid            bool                       `json:"valid"`
	Config           json.RawMessage            `json:"config"`
	OpenshiftObjects map[string]json.RawMessage `json:"openshiftObjects"`
}

type BooberReturn struct {
	Sources          json.RawMessage `json:"sources"`
	Errors           []string        `json:"errors"`
	Valid            bool            `json:"valid"`
	Config           json.RawMessage `json:"config"`
	OpenshiftObjects json.RawMessage `json:"openshiftObjects"`
}

func ExecuteSetup(args []string, dryRun bool, showConfig bool, showObjects bool, verbose bool, localhost bool,
	overrideFiles []string) (output string, error error) {

	var errorString string
	var affiliation string

	if !dryRun {
		if !validateLogin() {
			return "", errors.New("Not logged in, please use aoc login")
		}
		affiliation, error = GetAffiliation()
		if error != nil {
			return
		}
	}
	error= validateCommand(args, overrideFiles)
	if error != nil {
		return
	}

	var absolutePath string

	absolutePath, _ = filepath.Abs(args[0])

	var envFile string      // Filename for app
	var envFolder string    // Short folder name (Env)
	var folder string       // Absolute path of folder
	var parentFolder string // Absolute path of parent

	switch IsLegalFileFolder(args[0]) {
	case specIsFile:
		folder = filepath.Dir(absolutePath)
		envFile = filepath.Base(absolutePath)
	case specIsFolder:
		folder = absolutePath
		envFile = ""
	}

	parentFolder = filepath.Dir(folder)
	envFolder = filepath.Base(folder)

	if folder == parentFolder {
		errorString += fmt.Sprintf("Application configuration file cannot reside in root directory")
		return "", errors.New(errorString)
	}

	// Initialize JSON
	jsonStr, err := generateJson(envFile, envFolder, folder, parentFolder, args, overrideFiles, affiliation)
	if err != nil {
		return "", err
	} else {
		if dryRun {
			return fmt.Sprintf("%v", string(PrettyPrintJson(jsonStr))), nil
		} else {
			output, err = CallBoober(jsonStr, showConfig, showObjects, false, localhost, verbose)
			if err != nil {
				return "", err
			}
		}
	}
	return
}

// Check for valid login, that is we have a configuration with at least one reachable cluster
func validateLogin() bool {
	var openshiftCluster *openshift.OpenshiftCluster
	openshiftCluster = GetApiCluster()
	if openshiftCluster != nil {
		if !openshiftCluster.HasValidToken() {
			return false
		}
	}
	return true
}

func generateJson(envFile string, envFolder string, folder string, parentFolder string, args []string,
	overrideFiles []string, affiliation string) (jsonStr string, error error) {
	var booberData BooberInferface
	var returnMap map[string]json.RawMessage
	var returnMap2 map[string]json.RawMessage
	booberData.App = strings.TrimSuffix(envFile, filepath.Ext(envFile)) //envFile
	booberData.Env = envFolder

	booberData.Affiliation = affiliation

	returnMap, error = Folder2Map(folder, envFolder+"/")
	if error != nil {
		return
	}
	returnMap2, error = Folder2Map(parentFolder, "")
	if error != nil {
		return
	}

	booberData.Files = CombineMaps(returnMap, returnMap2)
	booberData.Overrides = overrides2map(args, overrideFiles)

	jsonByte, ok := json.Marshal(booberData)
	if !(ok == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling Boober data: %v\n", ok.Error()))
	}

	jsonStr = string(jsonByte)
	return
}

func overrides2map(args []string, overrideFiles []string) (returnMap map[string]json.RawMessage) {
	returnMap = make(map[string]json.RawMessage)
	for i := 0; i < len(overrideFiles); i++ {
		returnMap[overrideFiles[i]] = json.RawMessage(args[i+1])
	}
	return
}

func Folder2Map(folder string, prefix string) (returnMap map[string]json.RawMessage, error error) {
	returnMap = make(map[string]json.RawMessage)
	var allFilesOK bool = true
	var output string

	files, _ := ioutil.ReadDir(folder)
	var filesProcessed = 0
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if IsLegalFileFolder(absolutePath) == specIsFile { // Ignore folders
			matched, _ := filepath.Match("*.json", strings.ToLower(f.Name()))
			if matched {
				fileJson, err := ioutil.ReadFile(absolutePath)
				if err != nil {
					output += fmt.Sprintf("Error in reading file %v\n", absolutePath)
					allFilesOK = false
				} else {
					if IsLegalJson(string(fileJson)) {
						filesProcessed++
						returnMap[prefix+f.Name()] = fileJson
					} else {
						output += fmt.Sprintf("Illegal JSON in configuration file %v\n", absolutePath)
						allFilesOK = false
					}
				}
			}
		}

	}
	if !allFilesOK {
		error = errors.New(output)
	}
	return
}

func CombineMaps(map1 map[string]json.RawMessage, map2 map[string]json.RawMessage) (returnMap map[string]json.RawMessage) {
	returnMap = make(map[string]json.RawMessage)

	for k, v := range map1 {
		returnMap[k] = v
	}
	for k, v := range map2 {
		returnMap[k] = v
	}
	return
}

func validateCommand(args []string, overrideFiles []string) (error error) {
	var errorString = ""

	if len(args) == 0 {
		errorString += "Missing file/folder "
	} else {
		// Chceck argument 0 for legal file / folder
		validateCode := IsLegalFileFolder(args[0])
		if validateCode < 0 {
			errorString += fmt.Sprintf("Illegal file / folder: %v\n", args[0])
		}

		// We have at least one argument, now there should be a correlation between the number of args
		// and the number of override (-f) flags
		if len(overrideFiles) < (len(args) - 1) {
			errorString += fmt.Sprintf("Configuration override specified without file reference flag\n")
		}
		if len(overrideFiles) > (len(args) - 1) {
			errorString += fmt.Sprintf("Configuration overide file reference flag specified without configuration\n")
		}

		// Check for legal JSON argument for each overrideFiles flag
		for i := 1; i < len(args); i++ {
			if !IsLegalJson(args[i]) {
				errorString += fmt.Sprintf("Illegal JSON configuration override: %v\n", args[i])
			}
		}
	}

	if errorString != "" {
		error = errors.New(errorString)
	}
	return
}

func IsLegalFileFolder(filespec string) int {
	var err error
	var absolutePath string
	var fi os.FileInfo

	absolutePath, err = filepath.Abs(filespec)
	fi, err = os.Stat(absolutePath)
	if os.IsNotExist(err) {
		return specIllegal
	} else {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			return specIsFolder
		case mode.IsRegular():
			return specIsFile
		}
	}
	return specIllegal
}

func GetBooberAddress(clusterName string, localhost bool) (booberAddress string) {
	if localhost {
		booberAddress = "http://localhost:8080"
	} else {
		booberAddress = "http://boober-mfp-boober." + clusterName + ".paas.skead.no"
	}
	return
}

func GetBooberSetupUrl(clusterName string, localhost bool) string {
	return GetBooberAddress(clusterName, localhost) + "/setup"
}

func GetApiCluster() *openshift.OpenshiftCluster {
	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	openshiftConfig, err := openshift.LoadOrInitiateConfigFile(configLocation)
	if err != nil {
		fmt.Println("Error in loading OpenShift configuration")
		return nil
	}
	for i := range openshiftConfig.Clusters {
		if openshiftConfig.Clusters[i].Reachable {
			return openshiftConfig.Clusters[i]
		}
	}
	return nil
}

func GetAffiliation() (string, error) {
	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	openshiftConfig, err := openshift.LoadOrInitiateConfigFile(configLocation)
	if err != nil {
		return "", errors.New("Error in loading OpenShift configuration")
	}
	return openshiftConfig.Affiliation, nil
}

func CallBoober(combindedJson string, showConfig bool, showObjects bool, api bool, localhost bool, verbose bool) (string, error) {
	//var openshiftConfig *openshift.OpenshiftConfig
	var configLocation = viper.GetString("HOME") + "/.aoc.json"
	var output string

	if localhost {
		var token string = ""
		apiCluster := GetApiCluster()
		if apiCluster != nil {
			token = apiCluster.Token
		}
		out, err := CallBooberInstance(combindedJson, showConfig, showObjects, verbose,
			GetBooberSetupUrl("localhost", localhost), token)
		if err != nil {
			return out, err
		} else {
			output = out
		}
	} else {
		openshiftConfig, err := openshift.LoadOrInitiateConfigFile(configLocation)
		if err != nil {
			return "", errors.New("Error in loading OpenShift configuration")
		}

		var errorString string
		var newline string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					out, err := CallBooberInstance(combindedJson, showConfig, showObjects, verbose,
						GetBooberSetupUrl(openshiftConfig.Clusters[i].Name, localhost),
						openshiftConfig.Clusters[i].Token)
					if err == nil {
						output += fmt.Sprintf("%v\n", out)
					} else {
						if err.Error() != "" {
							errorString += newline + err.Error()
							newline = "\n"
						}
					}
				}
			}
		}
		if errorString != "" {
			return output, errors.New(errorString)
		}
	}
	return output, nil
}

func CallBooberInstance(combindedJson string, showConfig bool, showObjects bool, verbose bool, url string, token string) (string, error) {
	var output string

	if verbose {
		fmt.Print("Sending config to Boober at " + url + "... ")
	}

	var jsonStr = []byte(combindedJson)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Internal error in NewRequest: %v", err))
	}

	req.Header.Set("Authentication", "Bearer: "+token)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("FAIL.  Error connecting to Boober service")
		}
		return "", errors.New(fmt.Sprintf("Error connecting to the Boober service on %v: %v", url, err))
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		var errorstring string
		if strings.Contains(bodyStr, booberNotInstalledResponse) {
			errorstring = fmt.Sprintf("Boober not available on %v", url)
		} else {
			errorstring = fmt.Sprintf("Internal error")
		}
		if verbose {
			if strings.Contains(bodyStr, booberNotInstalledResponse) {
				fmt.Println("FAIL.  Boober not available")
			} else {
				fmt.Println("FAIL.  Internal error")
			}
		}
		return "", errors.New(fmt.Sprintf(errorstring))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	// Check return for error
	var booberReturn BooberReturn
	err = json.Unmarshal(body, &booberReturn)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error unmarshalling Boober return: %v\n", err.Error()))
	}

	if !(booberReturn.Valid) {
		fmt.Println("Error in configuration: ")
		for _, message := range booberReturn.Errors {
			fmt.Println("  " + message)
		}
	} else {
		if verbose {
			var booberReturnObjects BooberReturnObjects
			err = json.Unmarshal(body, &booberReturnObjects)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Error unmarshalling Boober return: %v\n", err.Error()))
			}
			var countMap map[string]int = make(map[string]int)
			for key := range booberReturnObjects.OpenshiftObjects {
				countMap[key]++
			}

			var space string
			var count int
			var out string
			for key := range countMap {
				out += fmt.Sprintf("%v%v: %v", space, key, countMap[key])
				space = "  "
				count++
			}
			if count > 0 {
				output := fmt.Sprintf("OK.  Objects: %v  (%v)", count, out)
				fmt.Println(output)
			}
		}
	}

	if showConfig {
		output += PrettyPrintJson(string(booberReturn.Config))
	}

	if showObjects {
		output += PrettyPrintJson(string(booberReturn.OpenshiftObjects))
	}

	return output, nil
}
