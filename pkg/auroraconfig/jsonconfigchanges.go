package auroraconfig

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
)

// RemoveEntry removes a value in an AuroraConfigFile on specified path
func jsonRemoveEntry(auroraConfigFile *AuroraConfigFile, pathParts []string) error {

	var jsoncontent map[string]interface{}
	// Unmarshal content from configfile
	if err := unmarshalJsonFile(auroraConfigFile, &jsoncontent); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and remove the entry
	if err := jsonRemoveEntryRecursive(&jsoncontent, pathParts); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into configfile
	if err := marshalJsonFile(auroraConfigFile, &jsoncontent); err != nil {
		return err
	}

	return nil
}

func jsonRemoveEntryRecursive(jsonContent *map[string]interface{}, pathParts []string) error {
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
		if err := jsonRemoveEntryRecursive(&subContent, restOfPath); err != nil {
			return err
		}
	}

	return nil
}

// SetValue sets a value in an AuroraConfigFile on specified path
func jsonSetValue(auroraConfigFile *AuroraConfigFile, pathParts []string, value string) error {

	var content map[string]interface{}
	// Unmarshal content from file
	if err := unmarshalJsonFile(auroraConfigFile, &content); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and set the value
	if err := jsonSetOrCreateRecursive(&content, pathParts, value); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into file
	if err := marshalJsonFile(auroraConfigFile, &content); err != nil {
		return err
	}

	return nil
}

func jsonSetOrCreateRecursive(content *map[string]interface{}, pathParts []string, value string) error {
	firstOfPath, err := validateAndGetFirstOfPath(pathParts)
	if err != nil {
		return err
	}

	if len(pathParts) == 1 {
		jsonSetFoundValue(content, firstOfPath, value)
	} else {
		restOfPath := pathParts[1:]
		_, ok := (*content)[firstOfPath].(map[string]interface{})
		if !ok {
			logrus.Debugf("No key %s found. Creating it.\n", firstOfPath)
			(*content)[firstOfPath] = make(map[string]interface{})
		}
		subContent := (*content)[firstOfPath].(map[string]interface{})
		if err := jsonSetOrCreateRecursive(&subContent, restOfPath, value); err != nil {
			return err
		}
	}

	return nil
}

func jsonSetFoundValue(content *map[string]interface{}, key string, value string) error {
	logrus.Debugf("Setting %s = %s\n", key, value)
	(*content)[key] = value
	return nil
}

func unmarshalJsonFile(auroraConfigFile *AuroraConfigFile, content *map[string]interface{}) error {
	if err := json.Unmarshal([]byte(auroraConfigFile.Contents), &content); err != nil {
		return err
	}
	return nil
}

func marshalJsonFile(auroraConfigFile *AuroraConfigFile, content *map[string]interface{}) error {
	prettyfile, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	auroraConfigFile.Contents = string(prettyfile) + "\n"
	return nil
}
