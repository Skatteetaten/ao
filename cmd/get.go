package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/client"

	"encoding/json"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/spf13/cobra"
)

var (
	flagJSON       bool
	flagAsList     bool
	flagNoDefaults bool
)

var (
	getCmd = &cobra.Command{
		Use:         "get",
		Short:       "Retrieves information from the AuroraConfig repository",
		Annotations: map[string]string{"type": "remote"},
	}

	getDeploymentsCmd = &cobra.Command{
		Use:   "all",
		Short: "Get all applicationIds",
		RunE:  PrintAll,
	}

	getAppsCmd = &cobra.Command{
		Use:     "app [applications]",
		Short:   "Get all applications or all envrionments for one or more applications",
		Aliases: []string{"apps"},
		RunE:    PrintApplications,
	}

	getEnvsCmd = &cobra.Command{
		Use:     "env [envirionments]",
		Short:   "Get all environments og all applications for one or more environments",
		Aliases: []string{"envs"},
		RunE:    PrintEnvironments,
	}

	getSpecCmd = &cobra.Command{
		Use:   "spec <applicationId>",
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

	getSpecCmd.Flags().BoolVarP(&flagNoDefaults, "no-defaults", "", false, "exclude default values from output")
	getSpecCmd.Flags().BoolVarP(&flagJSON, "json", "", false, "print deploy spec as json")
	getDeploymentsCmd.Flags().BoolVarP(&flagAsList, "list", "", false, "print ApplicationIds as a list")
}

func PrintAll(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	deployments := fileNames.GetApplicationIds()

	var header string
	var rows []string
	if flagAsList {
		sort.Strings(deployments)
		header = "APPLICATIONID"
		rows = deployments
	} else {
		header, rows = GetApplicationIdTable(deployments)
	}

	DefaultTablePrinter(header, rows, cmd.OutOrStdout())

	return nil
}

func PrintApplications(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(fileNames.GetApplications()) < 1 {
		return errors.New("No applications available")
	}
	if len(args) > 0 {
		return PrintDeploySpecTable(args, fuzzy.APP_FILTER, cmd, fileNames)
	}

	applications := fileNames.GetApplications()
	DefaultTablePrinter("APPLICATIONS", applications, cmd.OutOrStdout())
	return nil
}

func PrintEnvironments(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(fileNames.GetEnvironments()) < 1 {
		return errors.New("No environments available")
	}

	if len(args) > 0 {
		return PrintDeploySpecTable(args, fuzzy.ENV_FILTER, cmd, fileNames)
	}

	envrionments := fileNames.GetEnvironments()
	DefaultTablePrinter("ENVIRONMENTS", envrionments, cmd.OutOrStdout())
	return nil
}

func PrintDeploySpecTable(args []string, filter fuzzy.FilterMode, cmd *cobra.Command, fileNames client.FileNames) error {
	var selected []string
	for _, arg := range args {
		matches := fuzzy.FindAllDeploysFor(filter, arg, fileNames.GetApplicationIds())
		if len(matches) == 0 {
			return errors.Errorf("No matches for %s", arg)
		}
		selected = append(selected, matches...)
	}
	specs, err := DefaultApiClient.GetAuroraDeploySpec(selected, true)
	if err != nil {
		return err
	}
	header, rows := GetDeploySpecTable(specs)
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())
	return nil
}

func GetDeploySpecTable(specs []client.AuroraDeploySpec) (string, []string) {
	var rows []string
	header := "CLUSTER\tENVIRONMENT\tAPPLICATION\tVERSION\tREPLICAS\tTYPE\tDEPLOY STRATEGY"
	pattern := "%v\t%v\t%v\t%v\t%v\t%v\t%v"
	sort.Slice(specs, func(i, j int) bool {
		return strings.Compare(specs[i].Value("name").(string), specs[j].Value("name").(string)) != 1
	})
	for _, spec := range specs {
		row := fmt.Sprintf(
			pattern,
			spec.Value("cluster"),
			spec.Value("envName"),
			spec.Value("name"),
			spec.Value("version"),
			spec.Value("replicas"),
			spec.Value("type"),
			spec.Value("deployStrategy/type"),
		)
		rows = append(rows, row)
	}
	return header, rows
}

func PrintDeploySpec(cmd *cobra.Command, args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	matches := fuzzy.FindMatches(search, fileNames.GetApplicationIds(), false)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched than one file. Search must be more specific.\n%v", matches)
	}

	split := strings.Split(matches[0], "/")

	if !flagJSON {
		spec, err := DefaultApiClient.GetAuroraDeploySpecFormatted(split[0], split[1], !flagNoDefaults)
		if err != nil {
			return err
		}
		cmd.Println(spec)
		return nil
	}

	spec, err := DefaultApiClient.GetAuroraDeploySpec(matches, !flagNoDefaults)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	cmd.Println(string(data))
	return nil
}

func PrintFile(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
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

	matches := fuzzy.FindMatches(search, fileNames, true)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched than one file. Search must be more specific.\n%v", matches)
	}

	auroraConfigFile, err := DefaultApiClient.GetAuroraConfigFile(matches[0])
	if err != nil {
		return err
	}

	fmt.Println(auroraConfigFile.ToPrettyJson())
	return nil
}
