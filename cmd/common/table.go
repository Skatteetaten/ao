package common

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

func DefaultTablePrinter(table []string, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.TabIndent)
	for _, line := range table {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func SortedTable(header string, items []string) []string {
	sort.Strings(items)
	return append([]string{header}, items...)
}

func GetDeploymentTable(deployments []string) []string {
	sort.Strings(deployments)
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
