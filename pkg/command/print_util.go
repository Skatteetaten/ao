package command

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"os"
	"sort"
	"text/tabwriter"
)

func DefaultTablePrinter(lines []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func PrintDeployments(deployments []string) {
	sort.Strings(deployments)
	lines := GetDeploymentTable(deployments)
	DefaultTablePrinter(lines)
}

func PrintDeployResults(deploys []client.DeployResult) {
	results := []string{"\x1b[00mSTATUS\x1b[0m\tAPPLICATION\tENVIRONMENT\tCLUSTER\tDEPLOY_ID\t"}
	// TODO: Can we find the failed object?
	for _, item := range deploys {
		ads := item.ADS
		pattern := "%s\t%s\t%s\t%s\t%s\t"
		status := "\x1b[32mDeployed\x1b[0m"
		if !item.Success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, ads.Name, ads.Namespace, ads.Cluster, item.DeployId)
		results = append(results, result)
	}

	if len(deploys) > 0 {
		DefaultTablePrinter(results)
	}
}
