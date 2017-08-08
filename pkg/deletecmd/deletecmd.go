package deletecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"strings"
)

const UsageString = "Usage: aoc delete vault <vaultname> | secret <vaultname> <secretname> | app <appname> | env <envname> | deployment <envname> <appname> | file <filename>"
const vaultDontExistsError = "Error: No such vault"

type DeletecmdClass struct {
	configuration  configuration.ConfigurationClass
	deleteFileList []string
	force          bool
}

func (deletecmdClass *DeletecmdClass) getAffiliation() (affiliation string) {
	if deletecmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = deletecmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (deletecmdClass *DeletecmdClass) addDeleteFile(filename string) {
	deletecmdClass.deleteFileList = append(deletecmdClass.deleteFileList, filename)
}

func (deletecmdClass *DeletecmdClass) isFileDeleted(filename string) bool {
	for i := range deletecmdClass.deleteFileList {
		if deletecmdClass.deleteFileList[i] == filename {
			return true
		}
	}
	return false
}

func (deletecmdClass *DeletecmdClass) deleteFilesInList(auroraConfig serverapi_v2.AuroraConfig, persistentOptions *cmdoptions.CommonCommandOptions, affiliation string, openshiftConfig *openshift.OpenshiftConfig) error {
	// Delete all files in list
	for i := range deletecmdClass.deleteFileList {
		delete(auroraConfig.Files, deletecmdClass.deleteFileList[i])
	}
	return auroraconfig.PutAuroraConfig(auroraConfig, persistentOptions, affiliation, openshiftConfig)
}

func (deletecmdClass *DeletecmdClass) deleteVault(vaultName string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	//var vaults []serverapi_v2.Vault
	//vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
	_, err = auroraconfig.DeleteVault(vaultName, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}
	return
}

func (deletecmdClass *DeletecmdClass) deleteSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	//var vaults []serverapi_v2.Vault
	//vaults, err = auroraconfig.GetVaults(persistentOptions, createcmdClass.getAffiliation(), createcmdClass.configuration.GetOpenshiftConfig())
	fmt.Println("DEBUG: Delete secret called: " + vaultName + "/" + secretName)
	//vaults, err := auroraconfig.GetVaults(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())

	return
}

func (deletecmdClass *DeletecmdClass) addDeleteFileWithPrompt(filename string, prompt string) (err error) {
	if deletecmdClass.force {
		deletecmdClass.addDeleteFile(filename)
	} else {
		confirm, err := executil.PromptYNC(prompt)
		if err != nil {
			return err
		}
		if confirm == "Y" {
			deletecmdClass.addDeleteFile(filename)
		}
		if confirm == "C" {
			err = errors.New("Operation cancelled by user")
			return err
		}
	}
	return
}

func (deletecmdClass *DeletecmdClass) deleteApp(app string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/"+app+".json") {
			err = deletecmdClass.addDeleteFileWithPrompt(filename, "Delete file "+filename)
			if err != nil {
				return err
			}

			// Check if no other app in folder this app was deleted in.  If no more exists, then delete the env file about.json
			var parts []string = strings.Split(filename, "/")
			var env = parts[0]
			var otherAppDeployedInEnv bool = false
			for appfile := range auroraConfig.Files {
				if strings.Contains(appfile, env+"/") && !strings.Contains(appfile, "/about.json") {
					if !deletecmdClass.isFileDeleted(appfile) {
						otherAppDeployedInEnv = true
						break
					}
				}
			}

			if !otherAppDeployedInEnv {
				var aboutFile = env + "/about.json"
				err = deletecmdClass.addDeleteFileWithPrompt(aboutFile, "No other deployments in "+env+" exists, delete about file "+aboutFile)
				if err != nil {
					return err
				}
			}

		}
	}

	// Delete the root app file
	var rootAppFile string = app + ".json"
	err = deletecmdClass.addDeleteFileWithPrompt(rootAppFile, "Delete file "+rootAppFile)

	// Delete all files in list and update aurora config in boober
	err = deletecmdClass.deleteFilesInList(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) deleteEnv(env string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	// Delete all files in the folder
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, env+"/") {
			var parts []string = strings.Split(filename, "/")
			var app = parts[1]

			err = deletecmdClass.addDeleteFileWithPrompt(filename, "Delete file "+filename)
			if err != nil {
				return err
			}
			// If not about.json then check if this app is deployed in another env.  If not, then delete the root app.json also
			var deployAppFoundInOtherEnv bool = false
			if filename != env+"/about.json" {
				for appfile := range auroraConfig.Files {
					if strings.Contains(appfile, "/"+app) {
						// Check if file is marked for deletion, then we will not mark as found
						if !deletecmdClass.isFileDeleted(appfile) {
							deployAppFoundInOtherEnv = true
							break
						}
					}
				}
				if !deployAppFoundInOtherEnv {
					var rootFileName = app
					err = deletecmdClass.addDeleteFileWithPrompt(rootFileName, "No other deployment of "+rootFileName+" exists, delete root file "+rootFileName)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Delete all files in list and update aurora config in boober
	err = deletecmdClass.deleteFilesInList(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) deleteDeployment(env string, app string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	deploymentFilename := env + "/" + app + ".json"
	_, deploymentExists := auroraConfig.Files[deploymentFilename]
	if !deploymentExists {
		if !deletecmdClass.force {
			err = errors.New("No such deployment")
		}
		return err
	}

	err = deletecmdClass.addDeleteFileWithPrompt(deploymentFilename, "Delete file "+deploymentFilename)
	if err != nil {
		return err
	}

	// Check for existence of app.json in other envs, if none found then delete the root file as well
	var deployAppFoundInOtherEnv bool = false
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/"+app+".json") && filename != deploymentFilename {
			deployAppFoundInOtherEnv = true
			break
		}
	}

	if !deployAppFoundInOtherEnv {
		var rootFileName = app + ".json"
		err = deletecmdClass.addDeleteFileWithPrompt(rootFileName, "No other deployment of "+app+" exists, delete root file "+rootFileName)
		if err != nil {
			return err
		}
	}

	// Check for existence of other app.json in the same folder, if none found then delete the about.json as well
	var otherAppInSameFolder bool = false
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, env+"/") && strings.Contains(filename, ".json") && filename != deploymentFilename && filename != env+"/about.json" {
			otherAppInSameFolder = true
			break
		}
	}

	if !otherAppInSameFolder {
		var envAboutFilename string = env + "/about.json"
		err = deletecmdClass.addDeleteFileWithPrompt(envAboutFilename, "No other apps exists in the env "+env+", delete environment file "+envAboutFilename)
		if err != nil {
			return err
		}
	}

	// Delete all files in list and update aurora config in boober
	err = deletecmdClass.deleteFilesInList(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) deleteFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	_, deploymentExists := auroraConfig.Files[filename]
	if !deploymentExists {
		if !deletecmdClass.force {
			err = errors.New("No such file")
		}
		return err
	}

	deletecmdClass.addDeleteFile(filename)
	// Delete all files in list and update aurora config in boober
	err = deletecmdClass.deleteFilesInList(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) DeleteObject(args []string, force bool, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	err = validateDeletecmd(args)
	if err != nil {
		return
	}

	deletecmdClass.force = force

	var commandStr = args[0]
	switch commandStr {
	case "vault":
		{
			err = deletecmdClass.deleteVault(args[1], persistentOptions)
		}
	case "secret":
		{
			err = deletecmdClass.deleteSecret(args[1], args[2], persistentOptions)
		}
	case "app":
		{
			err = deletecmdClass.deleteApp(args[1], persistentOptions)
		}
	case "env":
		{
			err = deletecmdClass.deleteEnv(args[1], persistentOptions)
		}
	case "deployment":
		{
			err = deletecmdClass.deleteDeployment(args[1], args[2], persistentOptions)
		}
	case "file":
		{
			err = deletecmdClass.deleteFile(args[1], persistentOptions)
		}
	}
	return

}

func validateDeletecmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(UsageString)
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault", "app", "env", "file":
		{
			if len(args) != 2 {
				err = errors.New(UsageString)
				return
			}
		}
	case "secret", "deployment":
		{
			if len(args) != 3 {
				err = errors.New(UsageString)
				return
			}
		}
	}
	return

}
