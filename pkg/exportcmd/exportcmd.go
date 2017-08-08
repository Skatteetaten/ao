package exportcmd

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/jsonutil"
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
	configuration configuration.ConfigurationClass
}

func (exportcmdClass *ExportcmdClass) getAffiliation() (affiliation string) {
	if exportcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = exportcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

/*func (exportcmdClass *ExportcmdClass) exportVaults(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output, err = auroraconfig.GetVaults(persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return
	}
	return output, nil
} */

func (exportcmdClass *ExportcmdClass) exportVaults(outputFoldername string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output, err = auroraconfig.GetAllVaults(outputFoldername, persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return
	}
	if outputFoldername != "" {
		output = "Vaults are exported to " + outputFoldername
	}
	return output, err
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
			content, _, err := auroraconfig.GetContent(filename, persistentOptions, exportcmdClass.getAffiliation(), exportcmdClass.configuration.GetOpenshiftConfig())
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

	if len(args) > 1 {
		outputFolder = args[1]
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
	case "vaults":
		{
			output, err = exportcmdClass.exportVaults(outputFolder, persistentOptions)
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
