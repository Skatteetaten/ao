package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/cmd/common"
	"github.com/spf13/cobra"
	"strings"
)

var flagJson bool

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

	getSpecCmd.Flags().BoolVarP(&flagJson, "json", "", false, "print deploy spec as json")
}

func PrintAll(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	deployments := fileNames.GetDeployments()
	table := common.GetDeploymentTable(deployments)
	common.DefaultTablePrinter(table, cmd.OutOrStdout())

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

	table := common.SortedTable("APPLICATIONS", fileNames.GetApplications())
	common.DefaultTablePrinter(table, cmd.OutOrStdout())
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

	table := common.SortedTable("ENVIRONMENTS", fileNames.GetEnvironments())
	common.DefaultTablePrinter(table, cmd.OutOrStdout())
	return nil
}

func PrintDeploySpec(cmd *cobra.Command, args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}
	selected, err := common.SelectOne(args, fileNames.GetDeployments(), false)
	if err != nil {
		return err
	}

	split := strings.Split(selected, "/")

	if !flagJson {
		spec, err := DefaultApiClient.GetAuroraDeploySpecFormatted(split[0], split[1])
		if err != nil {
			return err
		}
		cmd.Println(spec)
		return nil
	}

	spec, err := DefaultApiClient.GetAuroraDeploySpec(split[0], split[1])
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
		table := common.GetFilesTable(fileNames)
		common.DefaultTablePrinter(table, cmd.OutOrStdout())
		return nil
	}

	selected, err := common.SelectOne(args, fileNames, true)
	if err != nil {
		return err
	}

	auroraConfigFile, err := DefaultApiClient.GetAuroraConfigFile(selected)
	if err != nil {
		return err
	}

	fmt.Println(auroraConfigFile.ToPrettyJson())
	return nil
}
