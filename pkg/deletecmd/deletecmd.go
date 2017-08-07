package deletecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/executil"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
	"strings"
)

const UsageString = "Usage: aoc delete vault <vaultname> | secret <vaultname> <secretname> | app <appname> | env <envname> | deployment <envname> <appname> | file <filename>"
const vaultDontExistsError = "Error: No such vault"

type DeletecmdClass struct {
	configuration  configuration.ConfigurationClass
	deleteFileList []string
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

func (deletecmdClass *DeletecmdClass) deleteFilesInList(auroraConfig serverapi_v2.AuroraConfig) (newAuroraConfig serverapi_v2.AuroraConfig, err error) {

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

// TODO: Delete all deployments of app in
// TODO: Chekc for existence of other app.json in all the envs affected, if none found then delete the root file as well
func (deletecmdClass *DeletecmdClass) deleteApp(app string, force bool, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	// Update aurora config in boober
	err = auroraconfig.PutAuroraConfig(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

// TODO: Check for existence of app.json for all apps in env in other envs, if none found then delete the root file as well
func (deletecmdClass *DeletecmdClass) deleteEnv(env string, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	// Update aurora config in boober
	err = auroraconfig.PutAuroraConfig(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) deleteDeployment(env string, app string, force bool, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	var deleteFileList []string

	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	deploymentFilename := env + "/" + app + ".json"
	_, deploymentExists := auroraConfig.Files[deploymentFilename]
	if !deploymentExists {
		if !force {
			err = errors.New("No such deployment")
		}
		return err
	}

	if force {
		deleteFileList = append(deleteFileList, deploymentFilename)
	} else {
		confirm, err := executil.PromptYNC("Delete file " + deploymentFilename)
		if err != nil {
			return err
		}
		if confirm == "Y" {
			deleteFileList = append(deleteFileList, deploymentFilename)
		}
		if confirm == "C" {
			return err
		}
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
		if force {
			deleteFileList = append(deleteFileList, app+".json")
		} else {
			confirm, err := executil.PromptYNC("No other deployment of " + app + " exists, delete root file " + app + ".json")
			if err != nil {
				return err
			}
			if confirm == "Y" {
				deleteFileList = append(deleteFileList, app+".json")
			}
			if confirm == "C" {
				return err
			}
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
		if force {
			deleteFileList = append(deleteFileList, env+"/about.json")
		} else {
			confirm, err := executil.PromptYNC("No other apps exists in the env " + env + ", delete environment file " + env + "/about.json")
			if err != nil {
				return err
			}
			if confirm == "Y" {
				deleteFileList = append(deleteFileList, env+"/about.json")
			}
			if confirm == "C" {
				return err
			}
		}
	}

	// Delete all files in list
	for i := range deleteFileList {
		delete(auroraConfig.Files, deleteFileList[i])
	}

	// Update aurora config in boober
	err = auroraconfig.PutAuroraConfig(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	return
}

func (deletecmdClass *DeletecmdClass) deleteFile(filename string, force bool, persistentOptions *cmdoptions.CommonCommandOptions) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
	if err != nil {
		return err
	}

	_, deploymentExists := auroraConfig.Files[filename]
	if !deploymentExists {
		if !force {
			err = errors.New("No such file")
		}
		return err
	}

	delete(auroraConfig.Files, filename)

	// Update aurora config in boober
	err = auroraconfig.PutAuroraConfig(auroraConfig, persistentOptions, deletecmdClass.getAffiliation(), deletecmdClass.configuration.GetOpenshiftConfig())
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

		}
	case "env":
		{
			//err = deletecmdClass.deleteEnv(args[1], persistentOptions)
		}
	case "deployment":
		{
			err = deletecmdClass.deleteDeployment(args[1], args[2], force, persistentOptions)
		}
	case "file":
		{
			err = deletecmdClass.deleteFile(args[1], force, persistentOptions)
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
