package auroraconfig

import (
	"errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// RemoveEntry removes a value in an AuroraConfigFile on specified path
func yamlRemoveEntry(auroraConfigFile *AuroraConfigFile, pathParts []string) error {

	var yamlcontent map[interface{}]interface{}
	// Unmarshal content from configfile
	if err := unmarshalYamlFile(auroraConfigFile, &yamlcontent); err != nil {
		return err
	}

	// Call the recursive parsing of content to locate and remove the entry
	if err := yamlRemoveEntryRecursive(&yamlcontent, pathParts); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into configfile
	if err := marshalYamlFile(auroraConfigFile, &yamlcontent); err != nil {
		return err
	}

	return nil
}

func yamlRemoveEntryRecursive(yamlContent *map[interface{}]interface{}, pathParts []string) error {
	firstOfPath, err := validateAndGetFirstOfPath(pathParts)
	if err != nil {
		return err
	}

	_, entryExists := (*yamlContent)[firstOfPath].(interface{})
	if !entryExists {
		return errors.New("No such path in target YAML document")
	}
	if len(pathParts) == 1 {
		delete((*yamlContent), firstOfPath)
	} else {
		restOfPath := pathParts[1:]
		subContent := (*yamlContent)[firstOfPath].(map[interface{}]interface{})
		if err := yamlRemoveEntryRecursive(&subContent, restOfPath); err != nil {
			return err
		}
	}

	return nil
}

func yamlSetValue(auroraConfigFile *AuroraConfigFile, pathParts []string, value string) error {

	var yamlcontent map[interface{}]interface{}
	// Unmarshal content from file
	if err := unmarshalYamlFile(auroraConfigFile, &yamlcontent); err != nil {
		return err
	}
	if len(yamlcontent) == 0 {
		yamlcontent = make(map[interface{}]interface{})
	}

	// Call the recursive parsing of content to locate and set the value
	if err := yamlSetOrCreateRecursive(&yamlcontent, pathParts, value); err != nil {
		return err
	}

	// Marshal changed content prettyfied back into file
	if err := marshalYamlFile(auroraConfigFile, &yamlcontent); err != nil {
		return err
	}

	return nil
}

func yamlSetOrCreateRecursive(content *map[interface{}]interface{}, pathParts []string, value string) error {
	firstOfPath, err := validateAndGetFirstOfPath(pathParts)
	if err != nil {
		return err
	}

	if len(pathParts) == 1 {
		yamlSetFoundValue(content, firstOfPath, value)
	} else {
		restOfPath := pathParts[1:]
		_, ok := (*content)[firstOfPath].(map[interface{}]interface{})
		if !ok {
			logrus.Debugf("No key %s found. Creating it.\n", firstOfPath)
			(*content)[firstOfPath] = make(map[interface{}]interface{})
		}
		subContent := (*content)[firstOfPath].(map[interface{}]interface{})
		if err := yamlSetOrCreateRecursive(&subContent, restOfPath, value); err != nil {
			return err
		}
	}

	return nil
}

func yamlSetFoundValue(content *map[interface{}]interface{}, key string, value string) error {
	logrus.Debugf("Setting %s = %s\n", key, value)
	(*content)[key] = value

	return nil
}

func unmarshalYamlFile(auroraConfigFile *AuroraConfigFile, content *map[interface{}]interface{}) error {
	if err := yaml.Unmarshal([]byte(auroraConfigFile.Contents), &content); err != nil {
		return err
	}

	return nil
}

func marshalYamlFile(auroraConfigFile *AuroraConfigFile, content *map[interface{}]interface{}) error {
	const yamlFileDashes = "---\n"
	prettyfile, err := yaml.Marshal(content)
	if err != nil {
		return err
	}
	auroraConfigFile.Contents = yamlFileDashes + string(prettyfile)

	return nil
}
