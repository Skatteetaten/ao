package editcmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/jsonutil"
)

const commentString = "# "
const editMessage = `
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
`

type EditcmdClass struct {
	Configuration *configuration.ConfigurationClass
}

func (editcmd *EditcmdClass) FuzzyEditFile(args []string) (string, error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs
	fuzzyArgs.Init(editcmd.Configuration)

	if err := fuzzyArgs.PopulateFuzzyEnvAppList(args[0:]); err != nil {
		return "", err
	}

	filename, err := fuzzyArgs.GetFile()
	if err != nil {
		return "", err
	}

	return editcmd.EditFile(filename)
}

func (editcmd *EditcmdClass) EditFile(filename string) (output string, err error) {

	var content string
	var version string

	content, version, err = auroraconfig.GetContent(filename, editcmd.Configuration)
	if err != nil {
		return "", err
	}
	content = jsonutil.PrettyPrintJson(content)

	var editCycleDone bool
	var modifiedContent = content

	for editCycleDone == false {
		contentBeforeEdit := modifiedContent
		modifiedContent, err = editString("# File: " + filename + editMessage + modifiedContent)
		if err != nil {
			return "", err
		}
		if (stripComments(modifiedContent) == stripComments(contentBeforeEdit)) || stripComments(modifiedContent) == stripComments(content) {
			if stripComments(modifiedContent) != stripComments(content) {
				tempfile, err := fileutil.CreateTempFile(stripComments(modifiedContent))
				if err != nil {
					return "", nil
				}
				output += "A copy of your changes har been stored to \"" + tempfile + "\"\n"
			}
			output += "Edit cancelled, no valid changes were saved."
			if editcmd.Configuration.PersistentOptions.Debug {
				fmt.Println("DEBUG: Content of modified file:")
				fmt.Println(modifiedContent)
				fmt.Println("DEBUG: Content of modified file stripped:")
				fmt.Println(stripComments(modifiedContent))
			}
			return output, nil
		}
		modifiedContent = stripComments(modifiedContent)

		if jsonutil.IsLegalJson(modifiedContent) {
			validationMessages, err := auroraconfig.PutFile(filename, modifiedContent, version, editcmd.Configuration)
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

func (editcmd *EditcmdClass) EditSecret(vaultName string, secretName string) (string, error) {

	secret, version, err := auroraconfig.GetSecret(vaultName, secretName, editcmd.Configuration)
	if err != nil {
		return "", err
	}

	var modifiedSecret = secret
	modifiedSecret, err = editString(modifiedSecret)
	if err != nil {
		return "", err
	}

	if modifiedSecret != secret {
		_, err = auroraconfig.PutSecret(vaultName, secretName, modifiedSecret, version, editcmd.Configuration)
	}

	return modifiedSecret, err
}

func addComments(content string, comments string) (commentedContent string, err error) {
	var commentLines []string

	const newline = "\n"
	var commentedComments string
	commentLines, _ = contentToLines(comments)
	for lineno := range commentLines {
		commentedComments += commentString + commentLines[lineno] + newline
	}
	commentedContent = commentedComments + content

	return
}

func stripComments(content string) (uncommentedContent string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var newline = ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "#") {
			uncommentedContent += newline + line
			newline = "\n"
		}
	}
	return
}

func contentToLines(content string) (contentLines []string, err error) {

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return
}

func editString(content string) (modifiedContent string, err error) {

	filename, err := fileutil.CreateTempFile(content)

	err = fileutil.EditFile(filename)
	if err != nil {
		return
	}

	fileText, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	modifiedContent = string(fileText)

	err = os.Remove(filename)
	if err != nil {
		err = errors.New("WARNING: Unable to delete tempfile " + filename)
	}
	return
}
