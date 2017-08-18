package editcmd

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

const usageString = "Usage: edit file [env/]<filename> | secret <vaultname> <secretname> "
const secretUseageString = "Usage: edit secret <vaultname> <secretname>"
const fileUseageString = "Usage: edit file [env/]<filename>"
const vaultUseageString = "Usage: edit vault <vaultname>"

const commentString = "# "
const editMessage = `
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
`

type EditcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (editcmd *EditcmdClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	editcmd.configuration.Init(persistentOptions)

	return
}

func (editcmd *EditcmdClass) EditSecret(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {

	var vaultname string = args[1]
	var secretname string = args[2]
	var version string = ""

	secret, version, err := auroraconfig.GetSecret(vaultname, secretname, &editcmd.configuration)
	if err != nil {
		return "", err
	}

	var modifiedSecret = secret
	modifiedSecret, err = editString(modifiedSecret)
	if err != nil {
		return "", err
	}

	if modifiedSecret != secret {
		_, err = auroraconfig.PutSecret(vaultname, secretname, modifiedSecret, version, &editcmd.configuration)
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

func (editcmd *EditcmdClass) EditObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	editcmd.init(persistentOptions)

	if !serverapi_v2.ValidateLogin(editcmd.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use ao login")
	}

	err = validateEditcmd(args)
	if err != nil {
		return
	}

	var fuzzyArgs fuzzyargs.FuzzyArgs
	err = fuzzyArgs.Init(&editcmd.configuration)
	if err != nil {
		return "", err
	}

	var commandStr = args[0]
	switch commandStr {
	case "file":
		{
			err = fuzzyArgs.PopulateFuzzyEnvAppList(args[1:])
			if err != nil {
				return "", err
			}
			filename, err := fuzzyArgs.GetFile()
			if err != nil {
				return "", err
			}
			output, err = editcmd.EditFile(filename, persistentOptions)
		}
	case "secret":
		{
			output, err = editcmd.EditSecret(args, persistentOptions)
		}
	case "vault":
		{
			output, err = editcmd.EditVault(args[1], persistentOptions)
		}
	default:
		{
			err = fuzzyArgs.PopulateFuzzyEnvAppList(args)
			if err != nil {
				return "", err
			}
			filename, err := fuzzyArgs.GetFile()
			if err != nil {
				return "", err
			}
			output, err = editcmd.EditFile(filename, persistentOptions)
		}
	}
	return

}

func validateEditcmd(args []string) (err error) {

	var commandStr = args[0]
	switch commandStr {
	case "file":
		{
			if len(args) < 2 {
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
	case "vault":
		{
			if len(args) < 2 {
				err = errors.New(vaultUseageString)
				return
			}
		}
	default:
		{
			// Might be a filename, just return and let the main proc parse it
			return
		}
	}
	return
}
