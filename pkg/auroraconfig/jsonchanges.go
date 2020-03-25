package auroraconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const pathSep = "/"

// SetValue sets a value in an AuroraConfigFile on specified path
func SetValue(auroraConfigFile *AuroraConfigFile, path string, value string) error {
	fmt.Printf("Got path: %s, value: %s\n", path, value)

	pathParts := getPathParts(path)
	if len(pathParts) == 0 {
		return errors.New("Too short path. No named key.")
	}

	var jsonContent map[string]interface{}
	if err := json.Unmarshal([]byte(auroraConfigFile.Contents), &jsonContent); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and set the value
	if err := setOrCreate(&jsonContent, pathParts, value); err != nil {
		return err
	}

	changedjson, err := json.Marshal(jsonContent)
	if err != nil {
		return err
	}
	auroraConfigFile.Contents = string(changedjson)

	return nil
}

func getPathParts(path string) []string {
	if path == "" {
		return nil
	}
	pathParts := strings.Split(path, pathSep)
	if strings.HasPrefix(path, pathSep) {
		pathParts = pathParts[1:]
	}
	if strings.HasSuffix(path, pathSep) {
		pathParts = pathParts[:len(pathParts)-1]
	}
	fmt.Printf("Got pathparts: %s\n", pathParts)
	return pathParts
}

func setOrCreate(jsonContent *map[string]interface{}, pathParts []string, value string) error {
	if len(pathParts) == 0 {
		return errors.New("No names in path")
	}

	fmt.Printf("Handling %s for value: %s\n", pathParts, value)
	fmt.Printf("--- jsonContent %s\n", jsonContent)

	firstOfPath := pathParts[0]
	fmt.Printf("--- firstOfPath: %s\n", firstOfPath)
	if len(pathParts) == 1 {
		setValue(jsonContent, firstOfPath, value)
	} else {
		restOfPath := pathParts[1:]
		_, ok := (*jsonContent)[firstOfPath].(map[string]interface{})
		if !ok {
			fmt.Printf("No key %s found. Creating it.\n", firstOfPath)
			(*jsonContent)[firstOfPath] = make(map[string]interface{})
		}
		subContent := (*jsonContent)[firstOfPath].(map[string]interface{})
		setOrCreate(&subContent, restOfPath, value)
	}

	return nil
}

func setValue(jsonContent *map[string]interface{}, key string, value string) error {
	fmt.Printf("Setting %s = %s\n", key, value)
	(*jsonContent)[key] = value
	return nil
}
