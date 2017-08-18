package editcmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

func (editcmd *EditcmdClass) EditFile(filename string) (output string, err error) {

	var content string
	var version string

	content, version, err = auroraconfig.GetContent(filename, editcmd.Configuration)
	if err != nil {
		return "", err
	}

	_, output, err = editCycle(content, filename, version, auroraconfig.PutFile, editcmd.Configuration)
	if err != nil {
		return "", err
	}
	return output, nil
}
