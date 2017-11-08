package editor

import (
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
	originalContent := PrettyPrintJson(content)
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

		err = Editor(tempFilePath)
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
		if currentContent == originalContent || beforeEdit.ModTime().Equal(afterEdit.ModTime()) {
			return "Edit cancelled, no changes made.", nil
		}

		if !IsLegalJson(currentContent) {
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

	return fmt.Sprintf("%s edited\n", fileName), nil
}
