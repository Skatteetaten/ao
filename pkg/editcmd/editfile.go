package editcmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

func (editcmd *EditcmdClass) EditFile(filename string) (output string, err error) {

	var content string
	var version string

	request := auroraconfig.GetAuroraConfigRequest(editcmd.Configuration)
	response, err := serverapi.CallApiWithRequest(request, editcmd.Configuration)
	if err != nil {
		return "", err
	}
	auroraConfig, err := auroraconfig.Response2AuroraConfig(response)
	if err != nil {
		return "", err
	}
	content, version, err = auroraconfig.GetContent(filename, &auroraConfig)
	if err != nil {
		return "", err
	}

	_, output, err = editCycle(content, filename, version, auroraconfig.PutFile, editcmd.Configuration)
	if err != nil {
		return "", err
	}
	return output, nil
}
