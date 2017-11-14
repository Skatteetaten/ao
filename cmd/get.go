package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves information from the AuroraConfig repository",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
}

var getDeploymentsCmd = &cobra.Command{
	Use:     "all",
	Short:   "Get all deployments",
	Long:    `Lists the deployments defined in the AuroraConfig`,
	Aliases: []string{"deployments"},
	RunE:    PrintDeployments,
}

var getAppsCmd = &cobra.Command{
	Use:     "app",
	Short:   "Get all applications",
	Long:    `Lists the apps defined in the Auroraconfig`,
	Aliases: []string{"apps"},
	RunE:    PrintApplications,
}

var getEnvsCmd = &cobra.Command{
	Use:     "env",
	Short:   "Get all environments",
	Long:    `Lists the envs defined in the Auroraconfig`,
	Aliases: []string{"envs"},
	RunE:    PrintEnvironments,
}

var getSpecCmd = &cobra.Command{
	Use:   "spec <search>",
	Short: "Get deploy spec for an application",
	Long:  `Deploy spec is the final result for merged AuroraConfig for one application`,
	RunE:  GetDeploySpec,
}

var getFileCmd = &cobra.Command{
	Use:   "files [envname]/<filename>",
	Short: "Get all files",
	Long: `Prints the content of the file to standard output.
Environmentnames and filenames can be abbrevated, and can be specified either as separate strings,
or on a env/file basis.

Given that a file called superapp-test/about.json exists in the repository, the command

	ao get file test/ab

will print the file.

If no argument is given, the command will list all the files in the repository.`,
	Aliases: []string{"file"},
	RunE:    PrintFile,
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getFileCmd)
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getEnvsCmd)
	getCmd.AddCommand(getDeploymentsCmd)
	getCmd.AddCommand(getSpecCmd)

	getSpecCmd.Flags().BoolP("json", "", false, "print deploy spec as json")
}

func GetDeploySpec(cmd *cobra.Command, args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	asJson, _ := cmd.Flags().GetBool("json")

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	matches, err := fuzzy.SearchForApplications(search, fileNames.GetDeployments())
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	}

	sort.Strings(matches)

	selected := matches[0]
	if len(matches) > 1 {
		selected = prompt.Select("Select application", matches)
	}

	if selected == "" {
		return nil
	}

	split := strings.Split(selected, "/")

	if !asJson {
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

func PrintDeployments(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	deployments := fileNames.GetDeployments()
	sort.Strings(deployments)
	table := GetDeploymentTable(deployments)
	DefaultTablePrinter(table, cmd.OutOrStdout())

	return nil
}

func PrintApplications(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	var table []string
	if len(args) > 0 {
		table = GetApplicationsTable(fileNames, args[0])
	} else {
		table = GetApplicationsTable(fileNames, "")
	}

	if len(table) < 2 {
		return errors.New("Did not find any applications")
	}

	DefaultTablePrinter(table, cmd.OutOrStdout())
	return nil
}

func PrintEnvironments(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	var table []string
	if len(args) > 0 {
		table = GetEnvironmentTable(fileNames, args[0])
	} else {
		table = GetEnvironmentTable(fileNames, "")
	}

	if len(table) < 2 {
		return errors.New("Did not find any environments")
	}

	DefaultTablePrinter(table, cmd.OutOrStdout())
	return nil
}

func PrintFile(cmd *cobra.Command, args []string) error {
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		table := GetFilesTable(fileNames)
		DefaultTablePrinter(table, cmd.OutOrStdout())
		return nil
	}

	file := args[0]
	matches, err := fuzzy.SearchForFile(file, fileNames)
	if err != nil {
		return err
	}

	if len(matches) < 1 {
		return errors.New("Did not find file " + file)
	}

	var selectedFile string
	if len(matches) == 1 {
		selectedFile = matches[0]
	} else {
		message := fmt.Sprintf("Matched %d files. Which file do you want?", len(matches))
		selectedFile = prompt.Select(message, matches)
	}

	auroraConfigFile, err := DefaultApiClient.GetAuroraConfigFile(selectedFile)
	if err != nil {
		return err
	}
	fmt.Println(auroraConfigFile.Name)
	fmt.Println(auroraConfigFile.ToPrettyJson())

	return nil
}

func DefaultTablePrinter(table []string, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.TabIndent)
	for _, line := range table {
		fmt.Fprintln(w, line)
	}
	w.Flush()
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
