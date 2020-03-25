package auroraconfig

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const pathSep = "/"

// RemoveEntry removes a value in an AuroraConfigFile on specified path
func RemoveEntry(auroraConfigFile *AuroraConfigFile, path string) error {
	pathParts := getPathParts(path)
	if len(pathParts) == 0 {
		return errors.New("Too short path. Must have a named key.")
	}

	// Unmarshal JSON content from file
	var jsonContent map[string]interface{}
	if err := json.Unmarshal([]byte(auroraConfigFile.Contents), &jsonContent); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and remove the entry
	if err := removeEntry(&jsonContent, pathParts); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into file
	prettyjson, err := json.MarshalIndent(jsonContent, "", "  ")
	if err != nil {
		return err
	}
	auroraConfigFile.Contents = string(prettyjson)

	return nil
}

func removeEntry(jsonContent *map[string]interface{}, pathParts []string) error {
	firstOfPath, err := validateAndGetFirstOfPath(pathParts)
	if err != nil {
		return err
	}

	_, entryExists := (*jsonContent)[firstOfPath].(interface{})
	if !entryExists {
		return errors.New("No such path in target JSON document")
	}
	if len(pathParts) == 1 {
		delete((*jsonContent), firstOfPath)
	} else {
		restOfPath := pathParts[1:]
		subContent := (*jsonContent)[firstOfPath].(map[string]interface{})
		if err := removeEntry(&subContent, restOfPath); err != nil {
			return err
		}
	}

	return nil
}

// SetValue sets a value in an AuroraConfigFile on specified path
func SetValue(auroraConfigFile *AuroraConfigFile, path string, value string) error {
	pathParts := getPathParts(path)
	if len(pathParts) == 0 {
		return errors.New("Too short path. Must have a named key.")
	}

	// Unmarshal JSON content from file
	var jsonContent map[string]interface{}
	if err := json.Unmarshal([]byte(auroraConfigFile.Contents), &jsonContent); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and set the value
	if err := setOrCreate(&jsonContent, pathParts, value); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into file
	prettyjson, err := json.MarshalIndent(jsonContent, "", "  ")
	if err != nil {
		return err
	}
	auroraConfigFile.Contents = string(prettyjson)

	return nil
}

func setOrCreate(jsonContent *map[string]interface{}, pathParts []string, value string) error {
	firstOfPath, err := validateAndGetFirstOfPath(pathParts)
	if err != nil {
		return err
	}

	if len(pathParts) == 1 {
		setValue(jsonContent, firstOfPath, value)
	} else {
		restOfPath := pathParts[1:]
		_, ok := (*jsonContent)[firstOfPath].(map[string]interface{})
		if !ok {
			logrus.Debugf("No key %s found. Creating it.\n", firstOfPath)
			(*jsonContent)[firstOfPath] = make(map[string]interface{})
		}
		subContent := (*jsonContent)[firstOfPath].(map[string]interface{})
		if err := setOrCreate(&subContent, restOfPath, value); err != nil {
			return err
		}
	}

	return nil
}

func setValue(jsonContent *map[string]interface{}, key string, value string) error {
	logrus.Debugf("Setting %s = %s\n", key, value)
	(*jsonContent)[key] = value
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
	return pathParts
}

func validateAndGetFirstOfPath(pathParts []string) (string, error) {
	if len(pathParts) == 0 {
		return "", errors.New("Path can not be empty")
	}
	firstOfPath := pathParts[0]
	isNumber, _ := regexp.MatchString(`^\d+$`, firstOfPath)
	if isNumber {
		return "", errors.New("Path can not have numeric entries")
	}
	return firstOfPath, nil
}
