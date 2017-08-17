package getcmd

import (
	"encoding/base64"
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

type GetcmdClass struct {
	Configuration *configuration.ConfigurationClass
}

func (getcmd *GetcmdClass) Files() (string, error) {
	var files []string
	files, err := auroraconfig.GetFileList(getcmd.Configuration)

	if err != nil {
		return "", err
	}

	output := "NAME"
	for fileindex := range files {
		output += "\n" + files[fileindex]
	}

	return output, nil
}

func (getcmd *GetcmdClass) File(args []string) (string, error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs
	if err := fuzzyArgs.Init(getcmd.Configuration); err != nil {
		return "", err
	}

	if err := fuzzyArgs.PopulateFuzzyEnvAppList(args); err != nil {
		return "", err
	}

	filename, err := fuzzyArgs.GetFile()
	if err != nil {
		return "", err
	}

	content, _, err := auroraconfig.GetContent(filename, getcmd.Configuration)
	if err != nil {
		return "", err
	}

	output := filename + ":\n"
	output += jsonutil.PrettyPrintJson(content)

	return output, err
}

func (getcmd *GetcmdClass) Clusters(persistentOptions *cmdoptions.CommonCommandOptions, clusterName string, allClusters bool) (string, error) {
	var displayClusterName string
	const tab = " "

	openshiftConfig := getcmd.Configuration.OpenshiftConfig
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

func (getcmd *GetcmdClass) Vaults() (string, error) {
	var vaults []serverapi_v2.Vault

	vaults, err := auroraconfig.GetVaultsArray(getcmd.Configuration)

	if err != nil {
		return "", err
	}

	output := "VAULT (Secrets)"
	for vaultindex := range vaults {
		numberOfSecrets := len(vaults[vaultindex].Secrets)
		output += "\n" + vaults[vaultindex].Name + " (" + fmt.Sprintf("%d", numberOfSecrets) + ")"
	}

	return output, err
}

func (getcmd *GetcmdClass) Vault(vaultName string) (string, error) {
	var vaults []serverapi_v2.Vault
	vaults, err := auroraconfig.GetVaultsArray(getcmd.Configuration)

	if err != nil {
		return "", err
	}

	output := "SECRET"
	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			for secretindex := range vaults[vaultindex].Secrets {
				output += "\n" + secretindex
			}
		}
	}

	return output, nil
}

func (getcmd *GetcmdClass) Secret(vaultName string, secretName string) (string, error) {
	var output string
	var vaults []serverapi_v2.Vault
	vaults, err := auroraconfig.GetVaultsArray(getcmd.Configuration)

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

func (getcmd *GetcmdClass) OcLogin(persistentOptions *cmdoptions.CommonCommandOptions) (string, error) {
	var kubeConfig kubernetes.KubeConfig
	cluster, user, token, err := kubeConfig.GetClusterUserAndToken()
	if err != nil {
		return "", err
	}

	output := "Cluster: " + cluster
	output += "\nUser: " + user
	output += "\nToken: " + token

	return output, nil
}
