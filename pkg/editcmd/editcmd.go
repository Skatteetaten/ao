package editcmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/serverapi"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const invalidConfigurationError = "Invalid configuration"
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
	err = validateEditcmd(args)
	if err != nil {
		return
	}
	if !serverapi.ValidateLogin(editcmdClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	var filename string = args[0]
	var content string

	content, err = editcmdClass.getContent(filename, persistentOptions)
	content = jsonutil.PrettyPrintJson(content)

	var editCycleDone bool
	var modifiedContent = content
	for editCycleDone == false {
		contentBeforeEdit := modifiedContent
		modifiedContent, err = editString(modifiedContent)
		if err != nil {
			return "", err
		}
		if (modifiedContent == contentBeforeEdit) || stripComments(modifiedContent) == content {
			if stripComments(modifiedContent) != content {
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
			validationMessages, err := editcmdClass.putContent(filename, modifiedContent, persistentOptions)
			if err != nil {
				if err.Error() == invalidConfigurationError {
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

func (editcmdClass *EditcmdClass) getContent(filename string, persistentOptions *cmdoptions.CommonCommandOptions) (content string, err error) {
	var affiliation = editcmdClass.getAffiliation()
	var apiEndpoint string = "/affiliation/" + affiliation + "/auroraconfig"
	var responses map[string]string
	var auroraConfig serverapi_v2.AuroraConfig

	responses, err = serverapi_v2.CallApi(http.MethodGet, apiEndpoint, "", persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, editcmdClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return "", err
			}
			if !response.Success {
				output, err := serverapi_v2.ResponsItems2MessageString(response)
				if err != nil {
					return "", err
				}
				return "", errors.New(output)

			}
		}

		return "", err
	}

	if len(responses) != 1 {
		err = errors.New("Internal error in getContent: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return "", err
		}
		auroraConfig, err = serverapi_v2.ResponseItems2AuroraConfig(response)

		var fileFound bool

		for filenameIndex := range auroraConfig.Files {
			if filenameIndex == filename {
				fileFound = true
				content = string(auroraConfig.Files[filenameIndex])
			}
		}
		if !fileFound {
			return "", errors.New("Illegal file/folder")
		}
	}

	return content, nil
}

func (editcmdClass *EditcmdClass) putContent(filename string, content string, persistentOptions *cmdoptions.CommonCommandOptions) (validationMessages string, err error) {
	var affiliation = editcmdClass.getAffiliation()
	var apiEndpoint = "/affiliation/" + affiliation + "/auroraconfig/" + filename
	var responses map[string]string
	responses, err = serverapi_v2.CallApi(http.MethodPut, apiEndpoint, content, persistentOptions.ShowConfig,
		persistentOptions.ShowObjects, true, persistentOptions.Localhost,
		persistentOptions.Verbose, editcmdClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
	if err != nil {
		for server := range responses {
			response, err := serverapi_v2.ParseResponse(responses[server])
			if err != nil {
				return "", err
			}
			if !response.Success {
				validationMessages, _ := serverapi_v2.ResponsItems2MessageString(response)
				return validationMessages, errors.New(invalidConfigurationError)
			}
		}

	}
	return
}

func editString(content string) (modifiedContent string, err error) {

	filename, err := createTempFile(editMessage + content)

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
		fmt.Println("WARNING: Unable to delete tempfile " + filename)
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

func validateEditcmd(args []string) (err error) {
	if len(args) != 1 {
		err = errors.New("Usage: aoc edit [env/]file")
		return
	}

	return
}
