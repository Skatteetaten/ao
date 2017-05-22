package editcmd

import (
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
)

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

	var affiliation = editcmdClass.getAffiliation()

	var filename string = args[0]
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
				output, err = serverapi_v2.ResponsItems2MessageString(response)
			}
		}
		return output, nil
	}

	if len(responses) != 1 {
		err = errors.New("Internal error: Response from " + strconv.Itoa(len(responses)))
		return
	}

	for server := range responses {
		response, err := serverapi_v2.ParseResponse(responses[server])
		if err != nil {
			return "", err
		}
		auroraConfig, err = serverapi_v2.ResponseItems2AuroraConfig(response)

		var fileFound bool
		var content string

		for filenameIndex := range auroraConfig.Files {
			if filenameIndex == filename {
				fileFound = true
				content = string(auroraConfig.Files[filenameIndex])
			}
		}
		if !fileFound {
			return "", errors.New("Illegal file/folder")
		}

		modifiedContent, err := editString(content)
		if err != nil {
			return "", err
		}

		apiEndpoint = "/affiliation/" + affiliation + "/auroraconfig/" + filename
		responses, err = serverapi_v2.CallApi(http.MethodPut, apiEndpoint, modifiedContent, persistentOptions.ShowConfig,
			persistentOptions.ShowObjects, true, persistentOptions.Localhost,
			persistentOptions.Verbose, editcmdClass.configuration.GetOpenshiftConfig(), persistentOptions.DryRun, persistentOptions.Debug)
		if err != nil {
			for server := range responses {
				response, err := serverapi_v2.ParseResponse(responses[server])
				if err != nil {
					return "", err
				}
				if !response.Success {
					output, err = serverapi_v2.ResponsItems2MessageString(response)
				}
			}
			return output, nil
		}

	}

	return
}

func editString(content string) (modifiedContent string, err error) {

	const tmpFilePrefix = ".aoc_edit_file_"
	var tmpDir = os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, tmpFilePrefix)
	if err != nil {
		return "", errors.New("Unable to create temporary file: " + err.Error())
	}
	fmt.Println("DEBUG: Temp file " + tmpFile.Name())
	if fileutil.IsLegalFileFolder(tmpFile.Name()) != fileutil.SpecIsFile {
		err = errors.New("Internal error: Illegal temp file name: " + tmpFile.Name())
	}
	err = ioutil.WriteFile(tmpFile.Name(), []byte(jsonutil.PrettyPrintJson(content)), 0700)
	if err != nil {
		return
	}

	fileutil.EditFile(tmpFile.Name())

	fileText, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return
	}
	modifiedContent = string(fileText)

	err = os.Remove(tmpFile.Name())
	if err != nil {
		fmt.Println("WARNING: Unable to delete tempfile " + tmpFile.Name())
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
