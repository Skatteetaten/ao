package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/command"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"sort"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves information from the AuroraConfig repository",
	Long:  `Can be used to retrieve one file or all the files from the respository.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var getDeploymentsCmd = &cobra.Command{
	Use:     "deployment",
	Short:   "get deployments",
	Long:    `Lists the deployments defined in the Auroraconfig`,
	Aliases: []string{"deployments", "dep", "deps", "all"},
	Run: func(cmd *cobra.Command, args []string) {

		fileNames, err := DefaultApiClient.GetFileNames()
		if err != nil {
			fmt.Println(err)
			return
		}

		deployments := fileNames.FilterDeployments()
		sort.Strings(deployments)
		table := command.GetDeploymentTable(deployments)
		command.DefaultTablePrinter(table)
	},
}

var getAppsCmd = &cobra.Command{
	Use:     "app",
	Short:   "get app",
	Long:    `Lists the apps defined in the Auroraconfig`,
	Aliases: []string{"apps"},
	Run: func(cmd *cobra.Command, args []string) {

		fileNames, err := DefaultApiClient.GetFileNames()
		if err != nil {
			fmt.Println(err)
			return
		}

		applications := fileNames.FilterDeployments()
		var table []string
		if len(args) > 0 {
			table = command.GetApplicationsTable(applications, args[0])
		} else {
			table = command.GetApplicationsTable(applications, "")
		}

		if len(table) < 2 {
			fmt.Println("Did not find any applications")
			return
		}

		command.DefaultTablePrinter(table)
	},
}

var getEnvsCmd = &cobra.Command{
	Use:     "env",
	Short:   "get env",
	Long:    `Lists the envs defined in the Auroraconfig`,
	Aliases: []string{"envs"},
	Run: func(cmd *cobra.Command, args []string) {

		fileNames, err := DefaultApiClient.GetFileNames()
		if err != nil {
			fmt.Println(err)
			return
		}

		applications := fileNames.FilterDeployments()
		var table []string
		if len(args) > 0 {
			table = command.GetEnvironmentTable(applications, args[0])
		} else {
			table = command.GetEnvironmentTable(applications, "")
		}

		if len(table) < 2 {
			fmt.Println("Did not find any environments")
			return
		}

		command.DefaultTablePrinter(table)
	},
}

var getFileCmd = &cobra.Command{
	Use:   "file [envname] <filename>",
	Short: "Get file",
	Long: `Prints the content of the file to standard output.
Environmentnames and filenames can be abbrevated, and can be specified either as separate strings,
or on a env/file basis.

Given that a file called superapp-test/about.json exists in the repository, the command

	ao get file test ab

will print the file.

If no argument is given, the command will list all the files in the repository.`,
	Aliases: []string{"files"},
	Run: func(cmd *cobra.Command, args []string) {

		fileNames, err := DefaultApiClient.GetFileNames()
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(args) < 1 {
			table := command.GetFilesTable(fileNames)
			command.DefaultTablePrinter(table)
			return
		}

		file := args[0]
		matches, err := fuzzy.SearchForFile(file, fileNames)
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(matches) < 1 {
			fmt.Println("Did not find file", file)
			return
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
			fmt.Println(err)
			return
		}
		fmt.Println(auroraConfigFile.ToPrettyJson())
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getFileCmd)
	getCmd.AddCommand(getAppsCmd)
	getCmd.AddCommand(getEnvsCmd)
	getCmd.AddCommand(getDeploymentsCmd)
}
