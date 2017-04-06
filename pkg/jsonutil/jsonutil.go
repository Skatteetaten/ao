package jsonutil

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const maxSecretFileSize int64 = 10 * 1024

// Struct to represent data to the Boober interface
type ApiInferface struct {
	Envs        []string                   `json:"envs"`
	Apps        []string                   `json:"apps"`
	Affiliation string                     `json:"affiliation"`
	Files       map[string]json.RawMessage `json:"files"`
	Overrides   map[string]json.RawMessage `json:"overrides"`
	SecretFiles map[string]string          `json:"secretFiles"`
}

func GenerateJson(envFile string, envFolder string, folder string, parentFolder string, overrideJson []string,
	overrideFiles []string, affiliation string) (jsonStr string, error error) {
	var apiData ApiInferface
	var returnMap map[string]json.RawMessage
	var returnMap2 map[string]json.RawMessage
	var secretMap map[string]string = make(map[string]string)

	apiData.Apps = make([]string, 1)
	apiData.Envs = make([]string, 1)
	apiData.Apps[0] = strings.TrimSuffix(envFile, filepath.Ext(envFile)) //envFile
	apiData.Envs[0] = envFolder

	apiData.Affiliation = affiliation

	returnMap, error = JsonFolder2Map(folder, envFolder+"/")
	if error != nil {
		return
	}

	returnMap2, error = JsonFolder2Map(parentFolder, "")
	if error != nil {
		return
	}

	apiData.Files = CombineJsonMaps(returnMap, returnMap2)
	apiData.Overrides = overrides2map(overrideJson, overrideFiles)
	apiData.SecretFiles = secretMap

	for fileKey := range apiData.Files {
		secret, err := json2secretFolder(apiData.Files[fileKey])
		if err != nil {
			return "", err
		}
		if secret != "" {
			secretMap, err = SecretFolder2Map(secret)
			if err != nil {
				return "", err
			}
			apiData.SecretFiles = CombineTextMaps(apiData.SecretFiles, secretMap)
		}
	}

	for overrideKey := range apiData.Overrides {
		secret, err := json2secretFolder(apiData.Overrides[overrideKey])
		if err != nil {
			return "", err
		}
		if secret != "" {
			secretMap, err = SecretFolder2Map(secret)
			if err != nil {
				return "", err
			}
			apiData.SecretFiles = CombineTextMaps(apiData.SecretFiles, secretMap)
		}
	}

	jsonByte, ok := json.Marshal(apiData)
	if !(ok == nil) {
		return "", errors.New(fmt.Sprintf("Internal error in marshalling Boober data: %v\n", ok.Error()))
	}

	jsonStr = string(jsonByte)
	return
}

// Search a json string for a secretFolder attribute
func json2secretFolder(jsonMessage json.RawMessage) (string, error) {
	type FileStruct struct {
		SecretFolder string `json:"secretFolder"`
	}
	var fileStruct FileStruct
	err := json.Unmarshal(jsonMessage, &fileStruct)
	if err != nil {
		return "", err
	}
	return fileStruct.SecretFolder, nil
}

func overrides2map(overrideJson []string, overrideFiles []string) (returnMap map[string]json.RawMessage) {
	returnMap = make(map[string]json.RawMessage)
	for i := 0; i < len(overrideFiles); i++ {
		returnMap[overrideFiles[i]] = json.RawMessage(overrideJson[i])
	}
	return
}

func JsonFolder2Map(folder string, prefix string) (map[string]json.RawMessage, error) {
	returnMap := make(map[string]json.RawMessage)
	var allFilesOK bool = true
	var output string
	files, _ := ioutil.ReadDir(folder)
	var filesProcessed = 0
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFile { // Ignore folders
			matched, _ := filepath.Match("*.json", strings.ToLower(f.Name()))
			if matched {
				fileJson, err := ioutil.ReadFile(absolutePath)
				if err != nil {
					output += fmt.Sprintf("Error in reading JSON file %v\n", absolutePath)
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
		return nil, errors.New(output + "Error in processing configuration files")
	}
	return returnMap, nil
}

func SecretFolder2Map(folder string) (map[string]string, error) {
	var allFilesOK bool = true
	var output string
	var returnMap map[string]string = make(map[string]string)

	files, _ := ioutil.ReadDir(folder)
	for _, f := range files {
		absolutePath := filepath.Join(folder, f.Name())
		if fileutil.IsLegalFileFolder(absolutePath) == fileutil.SpecIsFile { // Ignore folders

			fileSize, err := getFileSize(absolutePath)
			if err != nil {
				return nil, err
			}
			if fileSize > maxSecretFileSize {
				output += fmt.Sprintf("This secret is just too big: %v\n", absolutePath)
				allFilesOK = false
			}
			fileText, err := ioutil.ReadFile(absolutePath)
			fileTextBase64 := base64.StdEncoding.EncodeToString(fileText)
			if err != nil {
				output += fmt.Sprintf("Error in reading Secret file %v\n", absolutePath)
				allFilesOK = false
			} else {
				returnMap[absolutePath] = fileTextBase64
			}
		}
	}
	if !allFilesOK {
		return nil, errors.New(output + "Error in processing Secret files")
	}
	return returnMap, nil
}

func getFileSize(absolutePath string) (int64, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func CombineJsonMaps(map1 map[string]json.RawMessage, map2 map[string]json.RawMessage) (returnMap map[string]json.RawMessage) {
	returnMap = make(map[string]json.RawMessage)
	if map1 == nil {
		returnMap = map2
		return
	}
	if map2 == nil {
		returnMap = map1
		return
	}
	for k, v := range map1 {
		returnMap[k] = v
	}
	for k, v := range map2 {
		returnMap[k] = v
	}
	return
}

func CombineTextMaps(map1 map[string]string, map2 map[string]string) (returnMap map[string]string) {
	returnMap = make(map[string]string)
	if map1 == nil {
		returnMap = map2
		return
	}
	if map2 == nil {
		returnMap = map1
		return
	}
	for k, v := range map1 {
		returnMap[k] = v
	}
	for k, v := range map2 {
		returnMap[k] = v
	}
	return
}

func IsLegalJson(jsonString string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonString), &js) == nil
}

func PrettyPrintJson(jsonString string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(jsonString), "", "\t")
	if err != nil {
		return jsonString
	}
	return out.String()
}

func StripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}
