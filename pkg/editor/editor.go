package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const editPattern = `# Name: %s
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
%s%s
`

type OnSaveFunc func(modifiedContent string) ([]string, error)

func Edit(content string, fileName string, onSave OnSaveFunc) (string, error) {

	tempFilePath, err := CreateTempFile()
	if err != nil {
		return "", err
	}
	defer func() {
		err := os.Remove(tempFilePath)
		if err != nil {
			logrus.Fatal("WARNING: Unable to delete temp file ", tempFilePath)
		}
	}()

	var editErrors string
	originalContent := prettyPrintJson(content)
	currentContent := originalContent

	for {
		contentToEdit := fmt.Sprintf(editPattern, fileName, editErrors, currentContent)
		err = ioutil.WriteFile(tempFilePath, []byte(contentToEdit), 0700)
		if err != nil {
			return "", err
		}

		beforeEdit, err := os.Stat(tempFilePath)
		if err != nil {
			return "", err
		}

		err = OpenEditor(tempFilePath)
		if err != nil {
			return "", err
		}

		afterEdit, err := os.Stat(tempFilePath)
		if err != nil {
			return "", err
		}

		fileContent, err := ioutil.ReadFile(tempFilePath)
		if err != nil {
			return "", err
		}

		currentContent = stripComments(string(fileContent))

		originalHasChanges := HasContentChanged(originalContent, currentContent)
		noWrite := beforeEdit.ModTime().Equal(afterEdit.ModTime())
		if !originalHasChanges || noWrite {
			return "Edit cancelled, no changes made.", nil
		}

		if !json.Valid([]byte(currentContent)) {
			editErrors = addErrorMessage([]string{"Invalid JSON format"})
			continue
		}

		validationErrors, err := onSave(currentContent)
		if err != nil {
			return "", err
		} else if validationErrors != nil {
			editErrors = addErrorMessage(validationErrors)
		} else {
			// Content has been saved successfully
			break
		}
	}

	return fmt.Sprintf("%s edited", fileName), nil
}

func HasContentChanged(original, edited string) bool {

	orgBuffer := new(bytes.Buffer)
	err := json.Compact(orgBuffer, []byte(original))
	if err != nil {
		return true
	}

	editBuffer := new(bytes.Buffer)
	err = json.Compact(editBuffer, []byte(edited))
	if err != nil {
		return true
	}

	return orgBuffer.String() != editBuffer.String()
}
