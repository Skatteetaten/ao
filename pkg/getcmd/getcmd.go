package getcmd

import (
	"encoding/base64"
	"sort"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fuzzyargs"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/kubernetes"
	"github.com/skatteetaten/ao/pkg/printutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

type GetcmdClass struct {
	Configuration *configuration.ConfigurationClass
}

func (getcmd *GetcmdClass) Deployments(env string) (output string, err error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs

	err = fuzzyArgs.Init(getcmd.Configuration)
	if err != nil {
		return "", err
	}

	return formatDeploymentList(env, fuzzyArgs.LegalEnvList, fuzzyArgs.LegalDeployList), nil
}

func formatDeploymentList(envArg string, envList []string, appList []fuzzyargs.LegalDeployStruct) (output string) {

	var headers []string = []string{"ENV", "APP"}
	var envs []string
	var apps []string

	sort.Strings(envList)
	for _, env := range envList {
		if strings.Contains(env, envArg) || envArg == "" {
			var unsortedApps []string
			for i := range appList {
				if appList[i].EnvName == env {
					unsortedApps = append(unsortedApps, appList[i].AppName)
				}
			}
			sort.Strings(unsortedApps)
			var priorEnv string
			for _, app := range unsortedApps {
				if env != priorEnv {
					envs = append(envs, env)
					priorEnv = env
				} else {
					envs = append(envs, "")
				}
				apps = append(apps, app)
			}
		}
	}

	output = printutil.FormatTable(headers, envs, apps)

	return
}

func (getcmd *GetcmdClass) Apps() (output string, err error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs

	err = fuzzyArgs.Init(getcmd.Configuration)
	if err != nil {
		return "", err
	}
	return formatAppList(fuzzyArgs.LegalAppList), nil
}

func formatAppList(appList []string) (output string) {

	var headers []string = []string{"APP"}
	sort.Strings(appList)
	output = printutil.FormatTable(headers, appList)

	return
}

func (getcmd *GetcmdClass) Envs() (output string, err error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs

	err = fuzzyArgs.Init(getcmd.Configuration)
	if err != nil {
		return "", err
	}
	return formatEnvList(fuzzyArgs.LegalEnvList), nil
}

func formatEnvList(envList []string) (output string) {

	var headers []string = []string{"ENV"}
	sort.Strings(envList)
	output = printutil.FormatTable(headers, envList)

	return
}

func (getcmd *GetcmdClass) Files() (output string, err error) {
	auroraConfig, err := auroraconfig.GetAuroraConfig(getcmd.Configuration)
	if err != nil {
		return "", err
	}

	files := getFileList(&auroraConfig)
	if err != nil {
		return "", err
	}

	output = formatFileList(files)
	return output, nil
}

func formatFileList(files []string) (output string) {
	var headers []string = []string{"FILE/FOLDER", "FILE"}
	var parentCol []string
	var childCol []string

	sort.Strings(files)

	// List all files first
	for _, file := range files {
		if !strings.Contains(file, "/") {
			parentCol = append(parentCol, file)
			childCol = append(childCol, "")
		}
	}

	// Then list folders
	var previousParent string
	for _, file := range files {
		if strings.Contains(file, "/") {
			parts := strings.Split(file, "/")
			if previousParent != parts[0] {
				parentCol = append(parentCol, parts[0])
				previousParent = parts[0]
			} else {
				parentCol = append(parentCol, "")
			}
			childCol = append(childCol, parts[1])
		}
	}
	output = printutil.FormatTable(headers, parentCol, childCol)

	return output
}

func getFileList(auroraConfig *serverapi.AuroraConfig) (filenames []string) {

	filenames = make([]string, len(auroraConfig.Files))

	var filenameIndex = 0
	for filename := range auroraConfig.Files {
		filenames[filenameIndex] = filename
		filenameIndex++
	}
	return filenames
}

func (getcmd *GetcmdClass) File(args []string) (string, error) {
	var fuzzyArgs fuzzyargs.FuzzyArgs
	if err := fuzzyArgs.Init(getcmd.Configuration); err != nil {
		return "", err
	}

	if err := fuzzyArgs.PopulateFuzzyFile(args); err != nil {
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

func (getcmd *GetcmdClass) Clusters(clusterName string, allClusters bool) (output string, err error) {
	var displayClusterName string
	const tab = " "

	var headers []string = []string{"CLUSTER NAME", "REACHABLE", "LOGGED IN", "API", "URL"}
	var clusterNames []string
	var reachable []string
	var loggedIn []string
	var api []string
	var url []string

	openshiftConfig := getcmd.Configuration.OpenshiftConfig
	for i := range openshiftConfig.Clusters {
		if openshiftConfig.Clusters[i].Reachable || allClusters {
			displayClusterName = openshiftConfig.Clusters[i].Name
			if displayClusterName == clusterName || clusterName == "" {
				clusterNames = append(clusterNames, openshiftConfig.Clusters[i].Name)
				reachableColumn := ""
				if openshiftConfig.Clusters[i].Reachable {
					reachableColumn = "Yes"
				}
				reachable = append(reachable, reachableColumn)

				loggedInColumn := ""
				if openshiftConfig.Clusters[i].HasValidToken() {
					loggedInColumn = "Yes"
				}
				loggedIn = append(loggedIn, loggedInColumn)

				apiColumn := ""
				if openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					apiColumn = "Yes"
				}
				api = append(api, apiColumn)

				url = append(url, openshiftConfig.Clusters[i].Url)

			}
		}
	}

	output = printutil.FormatTable(headers, clusterNames, reachable, loggedIn, api, url)

	return output, nil
}

func (getcmd *GetcmdClass) Vaults(showSecretContent bool) (output string, err error) {
	var vaults []serverapi.Vault
	var headers []string = []string{"VAULT NAME", "SECRET NAME"}
	if showSecretContent {
		headers = append(headers, "SECRET CONTENT")
	}
	var sortedVaultNames []string
	var vaultNames []string
	var secretNames []string
	var secretContent []string

	vaults, err = auroraconfig.GetVaultsArray(getcmd.Configuration)

	if err != nil {
		return "", err
	}

	sortedVaultNames = make([]string, len(vaults))
	for i, vault := range vaults {
		sortedVaultNames[i] = vault.Name
	}

	sort.Strings(sortedVaultNames)
	for _, vaultName := range sortedVaultNames {
		vaultNames = append(vaultNames, vaultName)
		// Find all secrets in the vault
		for _, vault := range vaults {
			if vault.Name == vaultName {
				secretCount := 0
				for secretName := range vault.Secrets {
					secretNames = append(secretNames, secretName)
					if secretCount > 0 {
						vaultNames = append(vaultNames, "")
					}
					secretCount++
					if showSecretContent {
						decodedSecret, err := base64.StdEncoding.DecodeString(vault.Secrets[secretName])
						decodedSecretStr := string(decodedSecret)
						if strings.Contains(decodedSecretStr, "\n") {
							parts := strings.Split(decodedSecretStr, "\n")
							if len(parts) > 0 {
								decodedSecretStr = parts[0] + " ..."
							}
						}
						if err != nil {
							return "", err
						}
						secretContent = append(secretContent, decodedSecretStr)
					}
				}
			}
		}
	}

	if showSecretContent {
		output = printutil.FormatTable(headers, vaultNames, secretNames, secretContent)
	} else {
		output = printutil.FormatTable(headers, vaultNames, secretNames)
	}

	/*output := "VAULT (Secrets)"
	for vaultindex := range vaults {
		numberOfSecrets := len(vaults[vaultindex].Secrets)
		output += "\n" + vaults[vaultindex].Name + " (" + fmt.Sprintf("%d", numberOfSecrets) + ")"
	}*/

	return output, nil
}

func (getcmd *GetcmdClass) Vault(vaultName string) (string, error) {
	var vaults []serverapi.Vault
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
	var vaults []serverapi.Vault
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

func (getcmd *GetcmdClass) KubeConfig() (string, error) {
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

func (getcmd *GetcmdClass) OcLogin() (string, error) {
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
