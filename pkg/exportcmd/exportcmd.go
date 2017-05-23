package exportcmd

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
)

const UsageString = "Usage: aoc export files | file [env/]<filename> | adc"
const filesUsageString = "Usage: aoc export files"
const fileUseageString = "Usage: aoc export file [env/]<filename>"
const adcUsageString = "Usage: aoc export adc"
const notYetImplemented = "Not supported yet"

type ExportcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (exportcmdClass *ExportcmdClass) getAffiliation() (affiliation string) {
	if exportcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = exportcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (exportcmdClass *ExportcmdClass) exportFiles(outputFoldername string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	output, err = auroraconfig.GetAllContent(outputFoldername, persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return
	}
	if outputFoldername != "" {
		output = "Files are exported to " + outputFoldername
	}
	return output, err
}

func (exportcmdClass *ExportcmdClass) exportFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	switch outputFormat {
	case "json":
		{
			content, err := auroraconfig.GetContent(filename, persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.configuration.GetOpenshiftConfig())
			if err != nil {
				return "", err
			}
			output = jsonutil.PrettyPrintJson(content)
			return output, err
		}

	case "yaml":
		{
			err = errors.New(notYetImplemented)
			return "", err
		}
	default:
		{
			err = errors.New("Illegal format.  Legal values are json, yaml.")
		}
	}

	return
}

func (exportcmdClass *ExportcmdClass) getAdc(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output += notYetImplemented
	return
}

func (exportcmdClass *ExportcmdClass) ExportObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string, outputFolder string) (output string, err error) {
	err = validateExportcmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	if outputFormat == "" {
		outputFormat = "json"
	}

	switch commandStr {
	case "files":
		{
			output, err = exportcmdClass.exportFiles(outputFolder, persistentOptions, outputFormat)
		}
	case "file":
		{
			output, err = exportcmdClass.exportFile(args[1], persistentOptions, outputFormat)
		}
	case "adc":
		{
			output, err = exportcmdClass.getAdc(persistentOptions)
		}
	}

	return
}

func validateExportcmd(args []string) (err error) {
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
