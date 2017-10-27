package editcmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"os"
)

type OnSaveFunc = func(modifiedContent string) ([]string, error)

func editCycle(content string, contentName string, debug bool, onSave OnSaveFunc) (modifiedContent string, output string, err error) {

	var editCycleDone bool
	content = jsonutil.PrettyPrintJson(content)
	modifiedContent = content

	for !editCycleDone {
		contentBeforeEdit := modifiedContent
		modifiedContent, err = editString("# Name: " + contentName + editMessage + modifiedContent)
		if err != nil {
			return "", "", err
		}
		if (stripComments(modifiedContent) == stripComments(contentBeforeEdit)) || stripComments(modifiedContent) == stripComments(content) {
			if stripComments(modifiedContent) != stripComments(content) {
				tempfile, err := fileutil.CreateTempFile(stripComments(modifiedContent))
				if err != nil {
					return "", "", nil
				}
				output += "A copy of your changes har been stored to \"" + tempfile + "\"\n"
			}
			output += "Edit cancelled, no valid changes were saved."
			if debug {
				fmt.Println("DEBUG: Content of modified file:")
				fmt.Println(modifiedContent)
				fmt.Println("DEBUG: Content of modified file stripped:")
				fmt.Println(stripComments(modifiedContent))
			}
			return output, "", nil
		}
		modifiedContent = stripComments(modifiedContent)

		if jsonutil.IsLegalJson(modifiedContent) {
			validationMessages, err := onSave(modifiedContent)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if len(validationMessages) > 0{
				modifiedContent, _ = addComments(modifiedContent, validationMessages)
			} else {
				editCycleDone = true
			}
		} else {
			modifiedContent, _ = addComments(modifiedContent, []string{"Illegal JSON Format"})
		}
	}

	return
}
