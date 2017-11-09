package command

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"sort"
	"strings"
)

func GetVaultTable(vaults []*client.AuroraSecretVault) []string {
	table := []string{"VAULT\tPERMISSIONS\tSECRET\t"}

	sort.Slice(vaults, func(i, j int) bool {
		return strings.Compare(vaults[i].Name, vaults[j].Name) < 1
	})

	for _, vault := range vaults {
		name := vault.Name
		permissions := vault.Permissions.GetGroups()

		for s := range vault.Secrets {
			line := fmt.Sprintf("%s\t%s\t%s\t", name, permissions, s)
			table = append(table, line)
			name = " "
		}
	}

	return table
}

func GetFilesTable(files []string) []string {
	var single []string
	var envApp []string

	for _, file := range files {
		if strings.ContainsRune(file, '/') {
			envApp = append(envApp, file)
		} else {
			single = append(single, file)
		}
	}

	sort.Strings(single)
	sort.Strings(envApp)
	sortedFiles := append(single, envApp...)
	table := []string{"FILE"}
	for _, f := range sortedFiles {
		table = append(table, f)
	}

	return table
}

func GetDeploymentTable(deployments []string) []string {
	table := []string{"ENVIRONMENT\tAPPLICATION\t"}
	last := ""
	for _, app := range deployments {
		sp := strings.Split(app, "/")
		env := sp[0]
		app := sp[1]
		if env == last {
			env = " "
		}
		line := fmt.Sprintf("%s\t%s\t", env, app)
		table = append(table, line)
		last = sp[0]
	}

	return table
}

func GetEnvironmentTable(fileNames client.FileNames, env string) []string {
	return filterApplicationForTable(fuzzy.ENV_FILTER, fileNames, env)
}

func GetApplicationsTable(fileNames client.FileNames, app string) []string {
	return filterApplicationForTable(fuzzy.APP_FILTER, fileNames, app)
}

func filterApplicationForTable(mode fuzzy.FilterMode, fileNames client.FileNames, search string) []string {
	header := "ENVIRONMENT"
	if mode == fuzzy.APP_FILTER {
		header = "APPLICATION"
	}

	var matches []string
	if search != "" {
		matches = fuzzy.FindAllDeploysFor(mode, search, fileNames.GetDeployments())
		sort.Strings(matches)
		return GetDeploymentTable(matches)
	}

	table := []string{header}
	matches = fileNames.GetEnvironments()
	if mode == fuzzy.APP_FILTER {
		matches = fileNames.GetApplications()
	}

	sort.Strings(matches)
	for _, match := range matches {
		table = append(table, match)
	}

	return table
}
