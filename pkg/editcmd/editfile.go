package editcmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
)

func (editcmd *EditcmdClass) EditFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {

	var content string
	var version string

	content, version, err = auroraconfig.GetContent(filename, &editcmd.configuration)
	if err != nil {
		return "", err
	}

	_, output, err = editCycle(content, filename, version, auroraconfig.PutFile, &editcmd.configuration)
	if err != nil {
		return "", err
	}
	return output, nil
}
