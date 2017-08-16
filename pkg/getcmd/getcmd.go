package getcmd

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/kubernetes"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
)

const UsageString = "Usage: get files | vaults | vault <vaultname> | file [env/]<filename> | adc | secret <secretname> | cluster <clustername> | clusters | kubeconfig | oclogin"
const fileUseageString = "Usage: get file [[env/]<filename>]"
const vaultUseageString = "Usage: get vault [<vaultname>]"
const secretUseageString = "Usage: get secret <vaultname> <secretname>"

type GetcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (getcmd *GetcmdClass) init(persistentOptions *cmdoptions.CommonCommandOptions) (err error) {

	getcmd.configuration.Init(persistentOptions)
	return
}

func (getcmd *GetcmdClass) getFiles(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var files []string
	files, err = auroraconfig.GetFileList(&getcmd.configuration)

	output = "NAME"
	for fileindex := range files {
		output += "\n" + files[fileindex]
	}
	return
}

func (getcmd *GetcmdClass) getFile(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs
	err = fuzzyArgs.Init(&getcmd.configuration)
	if err != nil {
		return "", err
	}
	err = fuzzyArgs.PopulateFuzzyEnvAppList(args)
	if err != nil {
		return "", err
	}
	filename, err := fuzzyArgs.GetFile()
	if err != nil {
		return "", err
	}

	switch outputFormat {
	case "json":
		{
			content, _, err := auroraconfig.GetContent(filename, &getcmd.configuration)
			if err != nil {
				return "", err
			}
			output += filename + ":\n"
			output += jsonutil.PrettyPrintJson(content)
			return output, err
		}
	default:
		{
			err = errors.New("Illegal format: " + outputFormat + ".  Legal value are json.")
		}
	}

	return
}

func (getcmd *GetcmdClass) Clusters(persistentOptions *cmdoptions.CommonCommandOptions, clusterName string, allClusters bool) (string, error) {
	var displayClusterName string
	const tab = " "

	openshiftConfig := getcmd.configuration.GetOpenshiftConfig()
	output := "CLUSTER NAME         REACHABLE  LOGGED IN  API  URL"
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

	return output, nil
}

func (getcmd *GetcmdClass) getVaults(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var vaults []serverapi_v2.Vault

	vaults, err = auroraconfig.GetVaultsArray(&getcmd.configuration)

	output = "VAULT (Secrets)"
	for vaultindex := range vaults {
		numberOfSecrets := len(vaults[vaultindex].Secrets)
		output += "\n" + vaults[vaultindex].Name + " (" + fmt.Sprintf("%d", numberOfSecrets) + ")"
	}
	return
}

func (getcmd *GetcmdClass) getVault(vaultName string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
	var vaults []serverapi_v2.Vault
	vaults, err = auroraconfig.GetVaultsArray(&getcmd.configuration)

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

func (getcmd *GetcmdClass) Secret(vaultName string, secretName string, persistentOptions *cmdoptions.CommonCommandOptions) (string, error) {
	var output string
	var vaults []serverapi_v2.Vault
	vaults, err := auroraconfig.GetVaultsArray(&getcmd.configuration)

	if err != nil {
		return "", err
	}

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output += string(decodedSecret)
		}
	}
	return output, nil
}

func (getcmd *GetcmdClass) KubeConfig(persistentOptions *cmdoptions.CommonCommandOptions) (string, error) {
	var kubeConfig kubernetes.KubeConfig

	if err := kubeConfig.GetConfig(); err != nil {
		return "", err
	}

	output := "Current Context: " + kubeConfig.CurrentContext
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

	return output, nil
}

func (getcmd *GetcmdClass) OcLogin(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
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

func (getcmd *GetcmdClass) GetObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {
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
				output, err = getcmd.getFile(args[1:], persistentOptions, outputFormat)
			} else {
				output, err = getcmd.getFiles(persistentOptions)
			}
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
			if len(args) > 3 {
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
	default:
		{
			err = errors.New(UsageString)
			return
		}
	}

	return
}
