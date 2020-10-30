package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/deploymentspec"

	"encoding/json"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagJSON         bool
	flagAsList       bool
	flagNoDefaults   bool
	flagIgnoreErrors bool
)

var (
	getCmd = &cobra.Command{
		Use:         "get",
		Short:       "Retrieves information from the AuroraConfig repository",
		Annotations: map[string]string{"type": "remote"},
	}

	getDeploymentsCmd = &cobra.Command{
		Use:   "all",
		Short: "Get all applicationDeploymentRefs (environment/application)",
		RunE:  PrintAll,
	}

	getAppsCmd = &cobra.Command{
		Use:     "app [applications]",
		Short:   "Get all applications or all environments for one or more applications",
		Aliases: []string{"apps"},
		RunE:    PrintApplications,
	}

	getEnvsCmd = &cobra.Command{
		Use:     "env [environments]",
		Short:   "Get all environments or all applications for one or more environments",
		Aliases: []string{"envs"},
		RunE:    PrintEnvironments,
	}

	getSpecCmd = &cobra.Command{
		Use:   "spec <applicationDeploymentRef>",
		Short: "Get deploy spec for an application",
		RunE:  PrintDeploySpec,
	}

	getFileCmd = &cobra.Command{
		Use:     "file [environment/application]",
		Short:   "Get all files when no arguments are given or one specific file",
		Aliases: []string{"files"},
		RunE:    PrintFile,
	}
)

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getFileCmd)
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getEnvsCmd)
	getCmd.AddCommand(getDeploymentsCmd)
	getCmd.AddCommand(getSpecCmd)

	getSpecCmd.Flags().BoolVar(&flagNoDefaults, "no-defaults", false, "exclude default values from output")
	getSpecCmd.Flags().BoolVar(&flagJSON, "json", false, "print deploy spec as json")
	getSpecCmd.Flags().BoolVar(&flagIgnoreErrors, "ignore-errors", false, "suppresses errors from spec assembly. NB: may return incomplete deploy spec, use with care")
	getDeploymentsCmd.Flags().BoolVar(&flagAsList, "list", false, "print ApplicationDeploymentRefs as a list")
}

// PrintAll is the main method for the `get all` cli command
func PrintAll(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	deployments := fileNames.GetApplicationDeploymentRefs()

	var header string
	var rows []string
	if flagAsList {
		sort.Strings(deployments)
		header = "APPLICATIONDEPLOYMENTREF"
		rows = deployments
	} else {
		header, rows = GetApplicationDeploymentRefTable(deployments)
	}

	DefaultTablePrinter(header, rows, cmd.OutOrStdout())

	return nil
}

// PrintApplications is the main method for the `get app` cli command
func PrintApplications(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(fileNames.GetApplications()) < 1 {
		return errors.New("No applications available")
	}
	if len(args) > 0 {
		return PrintDeploySpecTable(args, auroraconfig.AppFilter, cmd, fileNames)
	}

	applications := fileNames.GetApplications()
	DefaultTablePrinter("APPLICATIONS", applications, cmd.OutOrStdout())
	return nil
}

// PrintEnvironments is the main method for the `get env` cli command
func PrintEnvironments(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(fileNames.GetEnvironments()) < 1 {
		return errors.New("No environments available")
	}

	if len(args) > 0 {
		return PrintDeploySpecTable(args, auroraconfig.EnvFilter, cmd, fileNames)
	}

	envrionments := fileNames.GetEnvironments()
	DefaultTablePrinter("ENVIRONMENTS", envrionments, cmd.OutOrStdout())
	return nil
}

// PrintDeploySpecTable prints a table of deployment specifications
func PrintDeploySpecTable(args []string, filter auroraconfig.FilterMode, cmd *cobra.Command, fileNames auroraconfig.FileNames) error {
	var selected []string
	for _, arg := range args {
		matches := auroraconfig.FindAllDeploysFor(filter, arg, fileNames.GetApplicationDeploymentRefs())
		if len(matches) == 0 {
			return errors.Errorf("No matches for %s", arg)
		}
		selected = append(selected, matches...)
	}
	specs, err := DefaultAPIClient.GetAuroraDeploySpec(selected, true, false)
	if err != nil {
		return err
	}
	header, rows := GetDeploySpecTable(specs, "")
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())
	return nil
}

// GetDeploySpecTable gets a table of deployment specifications
func GetDeploySpecTable(specs []deploymentspec.DeploymentSpec, newVersion string) (string, []string) {
	var rows []string
	releaseToDefined := false
	headers := []string{"CLUSTER", "ENVIRONMENT", "APPLICATION", "VERSION", "REPLICAS", "TYPE", "DEPLOY_STRATEGY"}
	sort.Slice(specs, func(i, j int) bool {
		return strings.Compare(specs[i].Name(), specs[j].Name()) != 1
	})

	for _, spec := range specs {
		if spec.HasValue("releaseTo") {
			headers = append(headers, "RELEASE_TO")
			releaseToDefined = true
			break
		}
	}

	pattern := makeColumnPattern(len(headers))

	for _, spec := range specs {
		var replicas string
		if spec.GetBool("pause") {
			replicas = "Paused"
		} else {
			replicas = fmt.Sprint(spec.GetString("replicas"))
		}
		specVersion := spec.Version()
		if newVersion != "" && len(specs) == 1 {
			specVersion = newVersion
		}
		specValues := []interface{}{spec.Cluster(), spec.Environment(), spec.Name(), specVersion, replicas, spec.GetString("type"), spec.GetString("deployStrategy/type")}
		if releaseToDefined {
			specValues = append(specValues, spec.GetString("releaseTo"))
		}

		row := fmt.Sprintf(
			pattern,
			specValues...,
		)
		rows = append(rows, row)
	}
	return strings.Join(headers, "\t"), rows
}

func makeColumnPattern(columnCount int) string {
	var slice []string
	for i := 0; i < columnCount; i++ {
		slice = append(slice, "%v")
	}
	pattern := strings.Join(slice, "\t")
	return pattern
}

// PrintDeploySpec is the main method for the `get spec` cli command
func PrintDeploySpec(cmd *cobra.Command, args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	matches := auroraconfig.FindMatches(search, fileNames.GetApplicationDeploymentRefs(), false)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched more than one file. Search must be more specific.\n%v", matches)
	}

	split := strings.Split(matches[0], "/")

	if !flagJSON {
		spec, err := DefaultAPIClient.GetAuroraDeploySpecFormatted(split[0], split[1], !flagNoDefaults, flagIgnoreErrors)
		if err != nil {
			return err
		}
		if flagIgnoreErrors {
			cmd.Println("NB: The following spec may be incomplete, since the ignore-errors flag was set.")
		}
		cmd.Println(spec)
		return nil
	}

	spec, err := DefaultAPIClient.GetAuroraDeploySpec(matches, !flagNoDefaults, flagIgnoreErrors)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	if flagIgnoreErrors {
		cmd.Println("NB: The following spec may be incomplete, since the ignore-errors flag was set.")
	}
	cmd.Println(string(data))
	return nil
}

// PrintFile is the main method for the `get file` cli command
func PrintFile(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		header, rows := GetFilesTable(fileNames)
		DefaultTablePrinter(header, rows, cmd.OutOrStdout())
		return nil
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	matches := auroraconfig.FindMatches(search, fileNames, true)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched more than one file. Search must be more specific.\n%v", matches)
	}

	auroraConfigFile, _, err := DefaultAPIClient.GetAuroraConfigFile(matches[0])
	if err != nil {
		return err
	}

	fmt.Println(auroraConfigFile.Contents)
	return nil
}
