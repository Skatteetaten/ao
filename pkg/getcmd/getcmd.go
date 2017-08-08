package getcmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/kubernetes"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

const UsageString = "Usage: get files | vaults | vault <vaultname> | file [env/]<filename> | adc | secret <secretname> | cluster <clustername> | clusters | kubeconfig | oclogin"
const filesUsageString = "Usage: get files"
const fileUseageString = "Usage: get file [[env/]<filename>]"
const vaultUseageString = "Usage: get vault [<vaultname>]"
const secretUseageString = "Usage: get secret <vaultname> <secretname>"
const adcUsageString = "Usage: get adc"
const notYetImplemented = "Not supported yet"

type GetcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (getcmd *GetcmdClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	getcmd.configuration.Init(persistentOptions)
	return
}

func (getcmd *GetcmdClass) getFiles(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var files []string
	files, err = auroraconfig.GetFileList(persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())

	output = "NAME"
	for fileindex := range files {
		output += "\n" + files[fileindex]
	}
	return
}

func (getcmd *GetcmdClass) getFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	switch outputFormat {
	case "json":
		{
			content, _, err := auroraconfig.GetContent(filename, persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())
			if err != nil {
				return "", err
			}
			output = jsonutil.PrettyPrintJson(content)
			return output, err
		}
	case "":
		{
			var files []string
			files, err = auroraconfig.GetFileList(persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())
			output = "NAME"
			var fileFound bool
			for fileindex := range files {
				if files[fileindex] == filename {
					output += "\n" + files[fileindex]
					fileFound = true
				}
			}
			if !fileFound {
				err = errors.New("Error: file \"" + filename + "\" not found")
				return "", err
			}
			return output, nil
		}
	case "yaml":
		{
			err = errors.New(notYetImplemented)
			return "", err
		}
	default:
		{
			err = errors.New("Illegal format: " + outputFormat + ".  Legal values are json, yaml.")
		}
	}

	return
}

func (getcmd *GetcmdClass) getAdc(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output += notYetImplemented
	return
}

func (getcmd *GetcmdClass) getClusters(persistentOptions *cmdoptions.CommonCommandOptions, clusterName string, allClusters bool) (output string, err error) {
	var displayClusterName string
	const tab = " "

	openshiftConfig := getcmd.configuration.GetOpenshiftConfig()
	output = "CLUSTER NAME         REACHABLE  LOGGED IN  API  URL"
	for i := range openshiftConfig.Clusters {
		if openshiftConfig.Clusters[i].Reachable || allClusters {
			displayClusterName = openshiftConfig.Clusters[i].Name
			if displayClusterName == clusterName || clusterName == "" {
				var apiColumn = fileutil.RightPad("", 4)
				if openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					apiColumn = fileutil.RightPad("Yes", 4)
				}
				var reachableColumn = fileutil.RightPad("", 10)
				if openshiftConfig.Clusters[i].Reachable {
					reachableColumn = fileutil.RightPad("Yes", 10)
				}
				var urlColumn = ""
				displayClusterName = fileutil.RightPad(displayClusterName, 20)
				urlColumn = openshiftConfig.Clusters[i].Url

				loggedInColumn := fileutil.RightPad("", 10)
				if openshiftConfig.Clusters[i].HasValidToken() {
					loggedInColumn = fileutil.RightPad("Yes", 10)
				}
				output += "\n" + displayClusterName + tab + reachableColumn + tab + loggedInColumn + tab + apiColumn + tab + urlColumn
			}
		}

	}

	return
}

func (getcmd *GetcmdClass) getVaults(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaultsArray(persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())

	output = "VAULT (Secrets)"
	for vaultindex := range vaults {
		numberOfSecrets := len(vaults[vaultindex].Secrets)
		output += "\n" + vaults[vaultindex].Name + " (" + fmt.Sprintf("%d", numberOfSecrets) + ")"
	}
	return
}

func (getcmd *GetcmdClass) getVault(vaultName string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaultsArray(persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			output = "SECRET"
			for secretindex := range vaults[vaultindex].Secrets {
				output += "\n" + secretindex
			}
		}

	}
	return
}

func (getcmd *GetcmdClass) getSecret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaultsArray(persistentOptions, getcmd.configuration.GetAffiliation(), getcmd.configuration.GetOpenshiftConfig())

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output += string(decodedSecret)
		}
	}
	return
}

func (getcmd *GetcmdClass) getKubeConfig(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var kubeConfig kubernetes.KubeConfig
	err = kubeConfig.GetConfig()
	if err != nil {
		return
	}

	output += "Current Context: " + kubeConfig.CurrentContext
	output += "\nClusters:"
	for i := range kubeConfig.Clusters {
		output += "\n\tName: " + kubeConfig.Clusters[i].Name
		output += "\n\t\tServer: " + kubeConfig.Clusters[i].Cluster.Server
	}
	output += "\nContexts:"
	for i := range kubeConfig.Contexts {
		output += "\n\tName: " + kubeConfig.Contexts[i].Name
		output += "\n\t\tCluster: " + kubeConfig.Contexts[i].Context.Cluster
		output += "\n\t\tNamespace: " + kubeConfig.Contexts[i].Context.Namespace
		output += "\n\t\tUser: " + kubeConfig.Contexts[i].Context.User
	}
	output += "\nUsers:"
	for i := range kubeConfig.Users {
		output += "\n\tName: " + kubeConfig.Users[i].Name
		output += "\n\t\tToken: " + kubeConfig.Users[i].User.Token
	}
	return
}

func (getcmd *GetcmdClass) getOcLogin(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var kubeConfig kubernetes.KubeConfig
	cluster, user, token, err := kubeConfig.GetClusterUserAndToken()
	if err != nil {
		return
	}
	output += "Cluster: " + cluster
	output += "\nUser: " + user
	output += "\nToken: " + token

	return
}

func (getcmd *GetcmdClass) GetObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string, allClusters bool) (output string, err error) {
	getcmd.init(persistentOptions)
	if !serverapi_v2.ValidateLogin(getcmd.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use ao login")
	}

	err = validateGetcmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "vault", "vaults":
		{
			if len(args) > 1 {
				output, err = getcmd.getVault(args[1], persistentOptions, outputFormat)
			} else {
				output, err = getcmd.getVaults(persistentOptions)
			}
		}
	case "file", "files":
		{
			if len(args) > 1 {
				output, err = getcmd.getFile(args[1], persistentOptions, outputFormat)
			} else {
				output, err = getcmd.getFiles(persistentOptions)
			}
		}
	case "secret":
		{
			output, err = getcmd.getSecret(args[1], args[2], persistentOptions, outputFormat)
		}
	case "adc":
		{
			output, err = getcmd.getAdc(persistentOptions)
		}
	case "cluster", "clusters":
		{
			var clusterName = ""
			if len(args) > 1 {
				clusterName = args[1]
			}
			output, err = getcmd.getClusters(persistentOptions, clusterName, allClusters)
		}
		// Deprecated when secrets are removed from AuroraConfig
		/*	case "secret", "secrets":
			{
				var secretName = ""
				if len(args) > 1 {
					secretName = args[1]
				}
				output, err = getcmdClass.getSecrets(persistentOptions, secretName)
			}*/
	case "kubeconfig":
		{
			output, err = getcmd.getKubeConfig(persistentOptions)
		}
	case "oclogin":
		{
			output, err = getcmd.getOcLogin(persistentOptions)
		}
	}

	return
}

func validateGetcmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(UsageString)
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "file", "files":
		{
			if len(args) > 2 {
				err = errors.New(fileUseageString)
				return
			}
		}
	case "vault", "vaults":
		{
			if len(args) > 2 {
				err = errors.New(vaultUseageString)
				return
			}
		}
	case "secret":
		{
			if len(args) != 3 {
				err = errors.New(secretUseageString)
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
	case "cluster", "clusters", "kubeconfig", "oclogin":
		{
			if len(args) > 1 {
				err = errors.New(UsageString)
			}
			return
		}
	default:
		{
			err = errors.New(UsageString)
			return
		}

	}

	return
}
