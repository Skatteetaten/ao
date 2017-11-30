package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

var (
	flagJson       bool
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
		Use:     "apps",
		Short:   "Get all applications",
		Aliases: []string{"app"},
		RunE:    PrintApplications,
	}

	getEnvsCmd = &cobra.Command{
		Use:     "envs",
		Short:   "Get all environments",
		Aliases: []string{"env"},
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
	getSpecCmd.Flags().BoolVarP(&flagJson, "json", "", false, "print deploy spec as json")
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

	applications := fileNames.GetApplications()
	sort.Strings(applications)
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

	envrionments := fileNames.GetEnvironments()
	sort.Strings(envrionments)
	DefaultTablePrinter("ENVIRONMENTS", envrionments, cmd.OutOrStdout())
	return nil
}

func PrintDeploySpec(cmd *cobra.Command, args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return cmd.Help()
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

	if !flagJson {
		spec, err := DefaultApiClient.GetAuroraDeploySpecFormatted(split[0], split[1], !flagNoDefaults)
		if err != nil {
			return err
		}
		cmd.Println(spec)
		return nil
	}

	spec, err := DefaultApiClient.GetAuroraDeploySpec(split[0], split[1], !flagNoDefaults)
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
