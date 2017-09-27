package exportcmd

import (
	"errors"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

const UsageString = "Usage: export files | file [env/]<filename> | vaults | adc"
const filesUsageString = "Usage: export files"
const fileUseageString = "Usage: export file [env/]<filename>"
const vaultsUsageString = "Usage: export vaults"
const vaultUsageString = "Usage: export vault <vaultname>"
const secretUsageString = "Usage: export secret <vaultname> <secretname>"
const adcUsageString = "Usage: export adc"
const notYetImplemented = "Not supported yet"

type ExportcmdClass struct {
	Configuration *configuration.ConfigurationClass
}

/*func (exportcmdClass *ExportcmdClass) exportVaults(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output, err = auroraconfig.GetVaults(persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.Configuration.GetOpenshiftConfig())
	if err != nil {
		return
	}
	return output, nil
} */

func (exportObj *ExportcmdClass) exportVaults(outputFoldername string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output, err = auroraconfig.GetAllVaults(outputFoldername, exportObj.Configuration)
	if err != nil {
		return
	}
	if outputFoldername != "" {
		output = "Vaults are exported to " + outputFoldername
	}
	return output, err
}

func (exportObj *ExportcmdClass) exportFiles(outputFoldername string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	output, err = auroraconfig.GetAllContent(outputFoldername, exportObj.Configuration)
	if err != nil {
		return
	}
	if outputFoldername != "" {
		output = "Files are exported to " + outputFoldername
	}
	return output, err
}

func (exportObj *ExportcmdClass) exportFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	switch outputFormat {
	case "json":
		{
			request := auroraconfig.GetAuroraConfigRequest(exportObj.Configuration)
			response, err := serverapi.CallApiWithRequest(request, exportObj.Configuration)
			if err != nil {
				return "", err
			}
			auroraConfig, err := auroraconfig.Response2AuroraConfig(response)
			if err != nil {
				return "", err
			}
			content, _, err := auroraconfig.GetContent(filename, &auroraConfig)
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

func (exportObj *ExportcmdClass) getAdc(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output += notYetImplemented
	return
}

func (exportObj *ExportcmdClass) ExportObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string, outputFolder string) (output string, err error) {
	err = validateExportcmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	if outputFormat == "" {
		outputFormat = "json"
	}

	if len(args) > 1 {
		outputFolder = args[1]
	}

	switch commandStr {
	case "files":
		{
			output, err = exportObj.exportFiles(outputFolder, persistentOptions, outputFormat)
		}
	case "file":
		{
			output, err = exportObj.exportFile(args[1], persistentOptions, outputFormat)
		}
	case "vaults":
		{
			output, err = exportObj.exportVaults(outputFolder, persistentOptions)
		}
	case "adc":
		{
			output, err = exportObj.getAdc(persistentOptions)
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
	case "vaults":
		{
			if len(args) > 1 {
				err = errors.New(vaultsUsageString)
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
	case "vault":
		{
			if len(args) != 2 {
				err = errors.New(vaultUsageString)
				return
			}
		}
	case "secret":
		{
			if len(args) != 2 {
				err = errors.New(secretUsageString)
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
