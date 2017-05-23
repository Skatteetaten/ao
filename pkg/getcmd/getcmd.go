package getcmd

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
)

const UsageString = "Usage: aoc get files | file [env/]<filename> | adc"
const filesUsageString = "Usage: aoc get files"
const fileUseageString = "Usage: aoc get file [env/]<filename>"
const adcUsageString = "Usage: aoc get adc"

type GetcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (getcmdClass *GetcmdClass) getAffiliation() (affiliation string) {
	if getcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = getcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (getcmdClass *GetcmdClass) GetObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}
	return
}

func validateEditcmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(UsageString)
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "files":
		{
			if len(args) > 1 {
				err = errors.New(filesUsageString)
				return
			}
		}
	case "file":
		{
			if len(args) != 2 {
				err = errors.New(fileUseageString)
				return
			}
		}
	case "adc":
		{
			if len(args) > 1 {
				err = errors.New(adcUsageString)
				return
			}
		}

	}

	return
}
