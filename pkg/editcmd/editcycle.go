package editcmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
)

type storeFunc func(string, string, string, *configuration.ConfigurationClass) (string, error)

func editCycle(content string, contentName string, version string, store storeFunc, configuration *configuration.ConfigurationClass) (modifiedContent string, output string, err error) {

	var editCycleDone bool
	content = jsonutil.PrettyPrintJson(content)
	modifiedContent = content

	for editCycleDone == false {
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
			if configuration.GetPersistentOptions().Debug {
				fmt.Println("DEBUG: Content of modified file:")
				fmt.Println(modifiedContent)
				fmt.Println("DEBUG: Content of modified file stripped:")
				fmt.Println(stripComments(modifiedContent))
			}
			return output, "", nil
		}
		modifiedContent = stripComments(modifiedContent)

		if jsonutil.IsLegalJson(modifiedContent) {
			validationMessages, err := store(modifiedContent, contentName, version, configuration)
			if err != nil {
				if err.Error() == auroraconfig.InvalidConfigurationError {
					modifiedContent, _ = addComments(modifiedContent, validationMessages)
				} else {
					editCycleDone = true
				}
			} else {
				editCycleDone = true
			}
		} else {
			modifiedContent, _ = addComments(modifiedContent, "Illegal JSON Format")
		}
	}

	return
}
