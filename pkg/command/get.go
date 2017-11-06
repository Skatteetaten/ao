package command

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"sort"
	"strings"
)

func GetAllDeploymentsTable(fileNames client.FileNames) []string {
	apps := fileNames.FilterDeployments()
	sort.Strings(apps)

	return GetDeploymentTable(apps)
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
