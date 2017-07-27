package editcmd

import (
	"bufio"
	"errors"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"io/ioutil"
	"os"
	"strings"
)

const usageString = "Usage: aoc edit file [env/]<filename> | secret <vaultname> <secretname> "
const secretUseageString = "Usage: aoc edit secret <vaultname> <secretname>"
const fileUseageString = "Usage: aoc edit file [env/]<filename>"

const commentString = "# "
const editMessage = `# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
`

type EditcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (editcmdClass *EditcmdClass) getAffiliation() (affiliation string) {
	if editcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = editcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (editcmdClass *EditcmdClass) EditFile(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {

	var filename string = args[1]
	var content string
	var version string

	content, version, err = auroraconfig.GetContent(filename, persistentOptions, editcmdClass.getAffiliation(), editcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return "", err
	}
	content = jsonutil.PrettyPrintJson(content)

	var editCycleDone bool
	var modifiedContent = content
	contentBeforeEdit := modifiedContent
	for editCycleDone == false {
		modifiedContent, err = editString(editMessage + modifiedContent)
		if err != nil {
			return "", err
		}
		if (stripComments(modifiedContent) == stripComments(contentBeforeEdit)) || stripComments(modifiedContent) == stripComments(content) {
			if stripComments(modifiedContent) != stripComments(content) {
				tempfile, err := createTempFile(stripComments(modifiedContent))
				if err != nil {
					return "", nil
				}
				output += "A copy of your changes har been stored to \"" + tempfile + "\"\n"
			}
			output += "Edit cancelled, no valid changes were saved."
			return output, nil
		}
		modifiedContent = stripComments(modifiedContent)

		if jsonutil.IsLegalJson(modifiedContent) {
			validationMessages, err := auroraconfig.PutFile(filename, modifiedContent, version, persistentOptions, editcmdClass.getAffiliation(), editcmdClass.configuration.GetOpenshiftConfig())
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

func (editcmdClass *EditcmdClass) EditSecret(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {

	var vaultname string = args[1]
	var secretname string = args[2]
	var version string = ""

	secret, version, err := auroraconfig.GetSecret(vaultname, secretname, persistentOptions, editcmdClass.getAffiliation(), editcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return "", err
	}

	var modifiedSecret = secret
	modifiedSecret, err = editString(modifiedSecret)
	if err != nil {
		return "", err
	}

	if modifiedSecret != secret {
		_, err = auroraconfig.PutSecret(vaultname, secretname, modifiedSecret, version, persistentOptions, editcmdClass.getAffiliation(), editcmdClass.configuration.GetOpenshiftConfig())
	}

	return
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
	var contenttLines []string

	var newline = ""
	contenttLines, _ = contentToLines(content)
	for lineno := range contenttLines {
		if strings.TrimLeft(contenttLines[lineno], commentString) == contenttLines[lineno] {
			uncommentedContent += newline + contenttLines[lineno]
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

	filename, err := createTempFile(content)

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

func createTempFile(content string) (filename string, err error) {
	const tmpFilePrefix = ".aoc_edit_file_"
	var tmpDir = os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, tmpFilePrefix)
	if err != nil {
		return "", errors.New("Unable to create temporary file: " + err.Error())
	}
	if fileutil.IsLegalFileFolder(tmpFile.Name()) != fileutil.SpecIsFile {
		err = errors.New("Internal error: Illegal temp file name: " + tmpFile.Name())
	}
	filename = tmpFile.Name()
	err = ioutil.WriteFile(tmpFile.Name(), []byte(content), 0700)
	if err != nil {
		return
	}
	return
}

func (editcmdClass *EditcmdClass) EditObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}
	if !serverapi.ValidateLogin(editcmdClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	var commandStr = args[0]
	switch commandStr {
	case "file":
		{
			output, err = editcmdClass.EditFile(args, persistentOptions)
		}
	case "secret":
		{
			output, err = editcmdClass.EditSecret(args, persistentOptions)
		}
	}
	return

}

func validateEditcmd(args []string) (err error) {

	var commandStr = args[0]
	switch commandStr {
	case "file":
		{
			if len(args) != 2 {
				err = errors.New(fileUseageString)
				return
			}
		}
	case "secret":
		{
			if len(args) != 3 {
				err = errors.New(secretUseageString)
				return
			}
		}
	default:
		{
			err = errors.New(usageString)
		}
	}
	return
}
