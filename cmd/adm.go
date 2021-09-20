package cmd

import (
	"fmt"
	"os"

	"github.com/skatteetaten/ao/pkg/versioncontrol"

	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var flagShowAll bool
var flagAddCluster []string
var flagBetaMultipleClusterTypes bool
var flagOnlyOcp3Clusters bool

var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Perform administrative commands on AO configuration",
}

var getClusterCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "List configured clusters",
	Aliases: []string{"cluster"},
	Run:     printClusters,
}

var getAffiliationCmd = &cobra.Command{
	Use:     "affiliations",
	Short:   "List defined affiliations",
	Aliases: []string{"affiliation"},
	RunE:    PrintAffiliationsGraphql,
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
	Use:   "update-hook <auroraconfig>",
	Short: `Update or create git hook to validate AuroraConfig.`,
	RunE:  UpdateGitHook,
}

var updateRefCmd = &cobra.Command{
	Use:   "update-ref <refName>",
	Short: `Update git ref for your auroraconfig checkout.`,
	RunE:  SetRefName,
}

var defaultApiClusterCmd = &cobra.Command{
	Use:   "default-apicluster <cluster>",
	Short: `Set configured default API cluster for ao.`,
	RunE:  SetApiCluster,
}

const (
	bashcompletionhelp = `
$ source ao.bash

# To load completions for each session, execute once:

$ sudo cp ao.bash /etc/bash_completion.d/ao
`
	zshcompletionhelp = `
# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit -u" >> ~/.zshrc

# To load completions for each session, execute once:

$ sudo cp ao.zsh "${fpath[1]}/_ao"

# You will need to start a new shell for this setup to take effect.
`
	fishcompletionhelp = `
$ source ao.fish

# To load completions for each session, execute once:
$ cp ao.fish ~/.config/fish/completions/ao.fish
`
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generates completion file for bash, zsh or fish. Default is bash",
	Long: `This command generates a shell script file that provides scripted completion for the specified shell.
Completion allows you to press the Tab key to complete keywords.
After running this command, a script file will exist in your directory.
To load completions:

Bash:
` + bashcompletionhelp +
		`
Zsh:
` + zshcompletionhelp +
		`
Fish:
` + fishcompletionhelp + `
`,
	RunE:      Completion,
	ValidArgs: []string{"bash", "zsh", "fish"},
}

func init() {
	RootCmd.AddCommand(admCmd)
	admCmd.AddCommand(getClusterCmd)
	admCmd.AddCommand(getAffiliationCmd)
	admCmd.AddCommand(completionCmd)
	admCmd.AddCommand(recreateConfigCmd)
	admCmd.AddCommand(updateClustersCmd)
	admCmd.AddCommand(updateHookCmd)
	admCmd.AddCommand(updateRefCmd)
	admCmd.AddCommand(defaultApiClusterCmd)

	getClusterCmd.Flags().BoolVarP(&flagShowAll, "all", "a", false, "Show all clusters, not just the reachable ones")
	recreateConfigCmd.Flags().BoolVarP(&flagBetaMultipleClusterTypes, "beta-multiple-cluster-types", "", false, "Generate new config for multiple cluster types. Eks ocp3, ocp4. (deprecated flag)")
	recreateConfigCmd.Flags().MarkHidden("beta-multiple-cluster-types")
	recreateConfigCmd.Flags().BoolVarP(&flagOnlyOcp3Clusters, "only-ocp3-clusters", "", false, "Generate new config for ocp3 only (deprecated function)")
	recreateConfigCmd.Flags().MarkHidden("only-ocp3-clusters")
	recreateConfigCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Recreate config with one cluster")
	recreateConfigCmd.Flags().StringArrayVarP(&flagAddCluster, "add-cluster", "a", []string{}, "Add cluster to available clusters")
	updateHookCmd.Flags().StringVarP(&flagGitHookType, "git-hook", "g", "pre-push", "Change git hook to validate AuroraConfig")
}

// PrintClusters is the main method for the `adm clusters` cli command
func PrintClusters(cmd *cobra.Command, printAll bool) {
	var rows []string
	for _, name := range AO.AvailableClusters {
		cluster := AO.Clusters[name]

		if !(cluster.Reachable || printAll) {
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

		apiURL := fmt.Sprintf("%s %s", cluster.BooberURL, cluster.GoboURL)

		api := ""
		if name == AO.APICluster {
			api = "Yes"
			if AO.Localhost {
				apiURL = "http://localhost:8080"
			}
		}
		line := fmt.Sprintf("\t%s\t%s\t%s\t%s\t%s\t%s", name, reachable, loggedIn, api, cluster.URL, apiURL)
		rows = append(rows, line)
	}

	header := "\tCLUSTER NAME\tREACHABLE\tLOGGED IN\tAPI\tURL\tAPI_URLS"
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())
}

func printClusters(cmd *cobra.Command, args []string) {
	PrintClusters(cmd, flagShowAll)
}

// SetRefName is the entry point for the `adm update-ref` cli command
func SetRefName(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Usage()
	}

	AO.RefName = args[0]
	if err := config.WriteConfig(*AO, ConfigLocation); err != nil {
		return err
	}

	cmd.Printf("refName = %s\n", args[0])
	return nil
}

// SetApiCluster is the entry point for the `adm default-apicluster` cli command
func SetApiCluster(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Usage()
	}

	newApiCluster := args[0]
	// Validate that newApiCluster is a valid cluster
	if _, ok := AO.Clusters[newApiCluster]; !ok {
		return fmt.Errorf("Failed to set default apicluster, %s is not a valid cluster", newApiCluster)
	}

	AO.APICluster = newApiCluster
	if err := config.WriteConfig(*AO, ConfigLocation); err != nil {
		return err
	}

	cmd.Printf("apiCluster = %s\n", newApiCluster)
	return nil
}

// UpdateClusters is the entry point for the `update-clusters` cli command
func UpdateClusters(cmd *cobra.Command, args []string) error {
	AO.InitClusters()
	AO.SelectAPICluster()
	return config.WriteConfig(*AO, ConfigLocation)
}

// RecreateConfig is the entry point for the `adm recreate-config` cli command
func RecreateConfig(cmd *cobra.Command, args []string) error {
	conf := &config.DefaultAOConfig
	if flagCluster != "" {
		conf.AvailableClusters = []string{flagCluster}
		conf.PreferredAPIClusters = []string{flagCluster}
	} else if len(flagAddCluster) > 0 {
		conf.AvailableClusters = append(conf.AvailableClusters, flagAddCluster...)
	}

	if flagBetaMultipleClusterTypes {
		fmt.Println("\nWarning: The flag --beta-multiple-cluster-types is redundant, since config for multiple cluster types is now default.")
	}

	if flagOnlyOcp3Clusters {
		fmt.Println("\nWarning: The flag --only-ocp3-clusters deletes access to ocp4 from config. It should only be used for ao development.")
	} else {
		conf.AddMultipleClusterConfig()
	}

	conf.InitClusters()
	conf.SelectAPICluster()
	return config.WriteConfig(*conf, ConfigLocation)
}

// Completion is the entry point for the `adm completion` cli command
func Completion(cmd *cobra.Command, args []string) error {
	shell := "bash"
	if len(args) > 1 {
		return cmd.Usage()
	} else if len(args) == 1 {
		shell = args[0]
	}
	var err error
	filename := ""
	helpfulInfo := ""
	switch shell {
	case "bash":
		filename = "ao.bash"
		err = RootCmd.GenBashCompletionFile(filename)
		helpfulInfo = bashcompletionhelp
	case "zsh":
		filename = "ao.zsh"
		err = RootCmd.GenZshCompletionFile(filename)
		helpfulInfo = zshcompletionhelp
	case "fish":
		filename = "ao.fish"
		err = RootCmd.GenFishCompletionFile(filename, true)
		helpfulInfo = fishcompletionhelp
	default:
		return cmd.Usage()
	}

	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	fmt.Println("Completion file created at", wd+"/"+filename)
	fmt.Println("To load completions:\n", helpfulInfo)
	return nil
}

// UpdateGitHook is the entry point for the `adm update-hook` cli command
func UpdateGitHook(cmd *cobra.Command, args []string) error {

	wd, _ := os.Getwd()
	gitPath, err := versioncontrol.FindGitPath(wd)
	if err != nil {
		return err
	}
	return versioncontrol.CreateGitValidateHook(gitPath, flagGitHookType, args[0])
}
