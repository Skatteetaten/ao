package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/skatteetaten/ao/pkg/versioncontrol"

	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var flagShowAll bool

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Perform administrative commands on AO configuration",
}

var getClusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "List configured clusters",
	Aliases: []string{"cluster"},
	Run:     PrintClusters,
}

var getAffiliationCmd = &cobra.Command{
	Use:     "affiliations",
	Short:   "List defined affiliations",
	Aliases: []string{"affiliation"},
	Run:     PrintAffiliations,
}
var updateClustersCmd = &cobra.Command{
	Use:   "update-clusters",
	Short: "Will recreate clusters in config file",
	RunE:  UpdateClusters,
}

var recreateConfigCmd = &cobra.Command{
	Use:   "recreate-config",
	Short: `The command will recreate the .ao.json file.`,
	RunE:  RecreateConfig,
}

var updateHookCmd = &cobra.Command{
	Use:   "update-hook",
	Short: `The command will recreate the .ao.json file.`,
	RunE:  UpdatePreCommitHook,
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion file",
	Long: `This command generates a bash script file that provides bash completion.
Bash completion allows you to press the Tab key to complete keywords.
After running this command, a file called ao.sh will exist in your home directory.
By typing
	source ./ao.sh
you will have bash completion in ao
To persist this across login sessions, please update your .bashrc file.`,
	RunE: BashCompletion,
}

func init() {
	RootCmd.AddCommand(admCmd)
	admCmd.AddCommand(getClusterCmd)
	admCmd.AddCommand(getAffiliationCmd)
	admCmd.AddCommand(completionCmd)
	admCmd.AddCommand(recreateConfigCmd)
	admCmd.AddCommand(updateClustersCmd)
	admCmd.AddCommand(updateHookCmd)

	getClusterCmd.Flags().BoolVarP(&flagShowAll, "all", "a", false, "Show all clusters, not just the reachable ones")
	recreateConfigCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Recreate config with one cluster")
}

func PrintClusters(cmd *cobra.Command, args []string) {
	var rows []string
	for _, name := range AO.AvailableClusters {
		cluster := AO.Clusters[name]

		if !(cluster.Reachable || flagShowAll) {
			continue
		}
		reachable := ""
		if cluster.Reachable {
			reachable = "Yes"
		}

		loggedIn := ""
		if cluster.HasValidToken() {
			loggedIn = "Yes"
		}

		api := ""
		if name == AO.APICluster {
			api = "Yes"
		}
		line := fmt.Sprintf("\t%s\t%s\t%s\t%s\t%s", name, reachable, loggedIn, api, cluster.Url)
		rows = append(rows, line)
	}

	header := "\tCLUSTER NAME\tREACHABLE\tLOGGED IN\tAPI\tURL"
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())
}

func PrintAffiliations(cmd *cobra.Command, args []string) {
	acn, err := DefaultApiClient.GetAuroraConfigNames()
	if err != nil {
		return
	}

	var mark string
	for _, affiliation := range *acn {
		if affiliation == AO.Affiliation {
			mark = "*"
		} else {
			mark = " "
		}
		line := fmt.Sprintf("  %s %s", mark, affiliation)
		cmd.Println(line)
	}
}

func UpdateClusters(cmd *cobra.Command, args []string) error {
	AO.InitClusters()
	AO.SelectApiCluster()
	return config.WriteConfig(*AO, ConfigLocation)
}

func RecreateConfig(cmd *cobra.Command, args []string) error {
	conf := &config.DefaultAOConfig
	if flagCluster != "" {
		conf.AvailableClusters = []string{flagCluster}
		conf.PreferredAPIClusters = []string{flagCluster}
	}
	conf.InitClusters()
	conf.SelectApiCluster()
	return config.WriteConfig(*conf, ConfigLocation)
}

func BashCompletion(cmd *cobra.Command, args []string) error {
	err := RootCmd.GenBashCompletionFile("ao.sh")
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	fmt.Println("Bash completion file created at", wd+"/ao.sh")
	return nil
}

func UpdatePreCommitHook(cmd *cobra.Command, args []string) error {
	wd, _ := os.Getwd()
	gitPath, found := versioncontrol.FindGitPath(wd)
	if !found {
		return errors.New("Could not find git")
	}
	return versioncontrol.CreatePreCommitHook(gitPath)
}
