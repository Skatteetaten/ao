package auroraconfig

import (
	"errors"
	"regexp"
	"strings"
)

const pathSep = "/"

// RemoveEntry removes a value in an AuroraConfigFile on specified path
func RemoveEntry(auroraConfigFile *File, path string) error {
	pathParts := getPathParts(path)
	if len(pathParts) == 0 {
		return errors.New("path is too short and must contain a named key")
	}

	if auroraConfigFile.IsYaml() {
		if err := yamlRemoveEntry(auroraConfigFile, pathParts); err != nil {
			return err
		}
	} else {
		if err := jsonRemoveEntry(auroraConfigFile, pathParts); err != nil {
			return err
		}
	}

	return nil
}

// SetValue sets a value in an AuroraConfigFile on specified path
func SetValue(auroraConfigFile *File, path string, value string) error {
	pathParts := getPathParts(path)
	if len(pathParts) == 0 {
		return errors.New("path is too short and must contain a named key")
	}

	if auroraConfigFile.IsYaml() {
		if err := yamlSetValue(auroraConfigFile, pathParts, value); err != nil {
			return err
		}
	} else {
		if err := jsonSetValue(auroraConfigFile, pathParts, value); err != nil {
			return err
		}
	}

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
