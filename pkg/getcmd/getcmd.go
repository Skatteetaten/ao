package getcmd

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/kubernetes"
)

const UsageString = "Usage: aoc get files | file [env/]<filename> | adc | secrets | secret <secretname> | cluster <clustername> | clusters | kubeconfig | oclogin"
const filesUsageString = "Usage: aoc get files"
const fileUseageString = "Usage: aoc get file [env/]<filename>"
const adcUsageString = "Usage: aoc get adc"
const notYetImplemented = "Not supported yet"

type GetcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (getcmdClass *GetcmdClass) getAffiliation() (affiliation string) {
	if getcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = getcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (getcmdClass *GetcmdClass) getFiles(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var files []string
	files, err = auroraconfig.GetFileList(persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())

	output = "NAME"
	for fileindex := range files {
		output += "\n" + files[fileindex]
	}
	return
}

func (getcmdClass *GetcmdClass) getFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	switch outputFormat {
	case "json":
		{
			content, _, err := auroraconfig.GetContent(filename, persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())
			if err != nil {
				return "", err
			}
			output = jsonutil.PrettyPrintJson(content)
			return output, err
		}
	case "":
		{
			var files []string
			files, err = auroraconfig.GetFileList(persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())
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
			err = errors.New("Illegal format.  Legal values are json, yaml.")
		}
	}

	return
}

func (getcmdClass *GetcmdClass) getAdc(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output += notYetImplemented
	return
}

func (getcmdClass *GetcmdClass) getClusters(persistentOptions *cmdoptions.CommonCommandOptions, clusterName string, allClusters bool) (output string, err error) {
	var displayClusterName string
	const tab = " "

	openshiftConfig := getcmdClass.configuration.GetOpenshiftConfig()
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

func (getcmdClass *GetcmdClass) getSecrets(persistentOptions *cmdoptions.CommonCommandOptions, secretName string) (output string, err error) {
	var secrets []string
	secrets, err = auroraconfig.GetSecretList(persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())

	output = "NAME"
	for secretindex := range secrets {
		if secretName == "" || secrets[secretindex] == secretName {
			output += "\n" + secrets[secretindex]
		}
	}
	return
}

func (getcmdClass *GetcmdClass) getKubeConfig(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
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

func (getcmdClass *GetcmdClass) getOcLogin(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
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

func (getcmdClass *GetcmdClass) GetObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string, allClusters bool) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "files":
		{
			output, err = getcmdClass.getFiles(persistentOptions)
		}
	case "file":
		{
			output, err = getcmdClass.getFile(args[1], persistentOptions, outputFormat)
		}
	case "adc":
		{
			output, err = getcmdClass.getAdc(persistentOptions)
		}
	case "cluster", "clusters":
		{
			var clusterName = ""
			if len(args) > 1 {
				clusterName = args[1]
			}
			output, err = getcmdClass.getClusters(persistentOptions, clusterName, allClusters)
		}
	case "secret", "secrets":
		{
			var secretName = ""
			if len(args) > 1 {
				secretName = args[1]
			}
			output, err = getcmdClass.getSecrets(persistentOptions, secretName)
		}
	case "kubeconfig":
		{
			output, err = getcmdClass.getKubeConfig(persistentOptions)
		}
	case "oclogin":
		{
			output, err = getcmdClass.getOcLogin(persistentOptions)
		}
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
	case "cluster", "clusters", "secret", "secrets", "kubeconfig", "oclogin":
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
