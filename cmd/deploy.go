package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/skatteetaten/ao/pkg/service"
	"github.com/spf13/cobra"
)

const deployLong = `Deploys applications from the current AuroraConfig.
For use in CI environments, use -y or --yes to disable interactivity and accept deployment.
`

const exampleDeploy = `  Given the following AuroraConfig:
    - about.json
    - foobar.json
    - bar.json
    - foo/about.json
    - foo/bar.json
		- foo/foobar.json
		- ref/about.json
    - ref/bar.json

  # Fuzzy matching: deploy foo/bar and foo/foobar
  ao deploy fo/ba

  # Exact matching: deploy foo/bar
  ao deploy foo/bar

  # Deploy an application with override for application file
  ao deploy foo/bar -o 'foo/bar.json:{"pause": true}'
	
  # Exclude application(s) from foo environment (regexp)
  ao deploy foo -e .*/bar -e .*/baz

  # Exclude environment(s) when deploying an application across environments (regexp)
  ao deploy bar -e ref/.*
`

var deployCmd = &cobra.Command{
	Aliases:     []string{"setup", "apply"},
	Use:         "deploy <applicationDeploymentRef>",
	Short:       "Deploy one or more ApplicationDeploymentRef (environment/application) to one or more clusters",
	Long:        deployLong,
	Example:     exampleDeploy,
	Annotations: map[string]string{"type": "actions"},
	RunE:        deploy,
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&flagAuroraConfig, "auroraconfig", "a", "", "Overrides the logged in AuroraConfig")
	deployCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Limit deploy to given cluster name")
	deployCmd.Flags().BoolVarP(&flagNoPrompt, "yes", "y", false, "Suppress prompts and accept deployment(s)")
	deployCmd.Flags().BoolVarP(&flagNoPrompt, "no-prompt", "", false, "Suppress prompts and accept deployment(s)")
	deployCmd.Flags().StringArrayVarP(&flagOverrides, "overrides", "o", []string{}, "Override in the form '[env/]file:{<json override>}'")
	deployCmd.Flags().StringArrayVarP(&flagExcludes, "exclude", "e", []string{}, "Select applications or environments to exclude from deploy")
	deployCmd.Flags().StringVarP(&flagVersion, "version", "v", "", "Set the given version in AuroraConfig before deploy")

	deployCmd.Flags().BoolVarP(&flagNoPrompt, "force", "f", false, "Suppress prompts and accept deployment(s)")
	deployCmd.Flags().MarkHidden("force")
	deployCmd.Flags().StringVarP(&flagAuroraConfig, "affiliation", "", "", "Overrides the logged in affiliation")
	deployCmd.Flags().MarkHidden("affiliation")
}

func deploy(cmd *cobra.Command, args []string) error {

	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	err := validateParams()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	auroraConfigName := AO.Affiliation
	if flagAuroraConfig != "" {
		auroraConfigName = flagAuroraConfig
	}

	apiClient, err := getAPIClient(auroraConfigName, pFlagToken, flagCluster)
	if err != nil {
		return err
	}

	applications, err := service.GetApplications(apiClient, search, flagExcludes)
	if err != nil {
		return err
	} else if len(applications) == 0 {
		return errors.New("No applications to deploy")
	}

	if flagVersion != "" && len(applications) > 1 {
		return errors.New("Deploy with version does only support one application")
	}

	filteredDeploymentSpecs, err := service.GetFilteredDeploymentSpecs(apiClient, applications, flagCluster)
	if err != nil {
		return err
	}

	overrideConfig, err := parseOverride(flagOverrides)
	if err != nil {
		return err
	}

	partitions, err := createDeploySpecPartitions(auroraConfigName, pFlagToken, AO.Clusters, filteredDeploymentSpecs)
	if err != nil {
		return err
	}

	if !getDeployConfirmation(flagNoPrompt, filteredDeploymentSpecs, flagVersion, cmd.OutOrStdout()) {
		return errors.New("No applications to deploy")
	}

	if flagVersion != "" {
		err = updateVersion(apiClient, applications, flagVersion, cmd.OutOrStdout())
		if err != nil {
			return err
		}
	}

	result, err := deployToReachableClusters(getApplicationDeploymentClient, partitions, overrideConfig)
	if err != nil {
		return err
	}

	printDeployResult(result, cmd.OutOrStdout())

	return nil
}

func validateParams() error {

	if flagCluster != "" {
		if _, exists := AO.Clusters[flagCluster]; !exists {
			return errors.New(fmt.Sprintf("No such cluster %s", flagCluster))
		}
	}

	return nil
}

func parseOverride(overrides []string) (map[string]string, error) {
	returnMap := make(map[string]string)
	for i := 0; i < len(overrides); i++ {
		indexByte := strings.IndexByte(overrides[i], ':')
		filename := overrides[i][:indexByte]
		jsonOverride := overrides[i][indexByte+1:]

		if !json.Valid([]byte(jsonOverride)) {
			msg := fmt.Sprintf("%s is not a valid json", jsonOverride)
			return nil, errors.New(msg)
		}

		returnMap[filename] = jsonOverride
	}
	return returnMap, nil
}

func getDeployConfirmation(force bool, filteredDeploymentSpecs []deploymentspec.DeploymentSpec, newVersion string, out io.Writer) bool {
	header, rows := GetDeploySpecTable(filteredDeploymentSpecs, newVersion)
	DefaultTablePrinter(header, rows, out)

	shouldDeploy := true
	if !force {
		defaultAnswer := len(rows) == 1
		message := fmt.Sprintf("Do you want to deploy %d application(s)?", len(rows))
		shouldDeploy = prompt.Confirm(message, defaultAnswer)
	}

	return shouldDeploy
}

func deployToReachableClusters(getClient func(partition Partition) client.ApplicationDeploymentClient, partitions []DeploySpecPartition, overrideConfig map[string]string) ([]client.DeployResults, error) {
	deployResult := make(chan client.DeployResults)

	for _, partition := range partitions {
		go performDeploy(getClient(partition.Partition), partition, overrideConfig, deployResult)
	}

	var allResults []client.DeployResults
	for i := 0; i < len(partitions); i++ {
		allResults = append(allResults, <-deployResult)
	}

	return allResults, nil
}

func performDeploy(deployClient client.ApplicationDeploymentClient, partition DeploySpecPartition, overrideConfig map[string]string, deployResults chan<- client.DeployResults) {
	if !partition.Cluster.Reachable {
		deployResults <- errorDeployResults("Cluster is not reachable", partition)
		return
	}

	var applicationList []string
	for _, spec := range partition.DeploySpecs {
		applicationList = append(applicationList, spec.GetString("applicationDeploymentRef"))
	}

	payload := client.NewDeployPayload(applicationList, overrideConfig)

	result, err := deployClient.Deploy(payload)
	if err != nil {
		deployResults <- errorDeployResults(err.Error(), partition)
	} else {
		deployResults <- *result
	}
}

func errorDeployResults(reason string, partition DeploySpecPartition) client.DeployResults {
	var applicationResults []client.DeployResult

	for _, spec := range partition.DeploySpecs {
		affiliation := spec.GetString("affiliation")
		applicationDeploymentRef := client.NewApplicationDeploymentRef(spec.GetString("applicationDeploymentRef"))

		result := new(client.DeployResult)
		result.DeployID = "-"
		result.Ignored = false
		result.Success = false
		result.Reason = reason
		result.DeploymentSpec = deploymentspec.NewDeploymentSpec(
			applicationDeploymentRef.Application,
			affiliation+"-"+applicationDeploymentRef.Environment,
			partition.Cluster.Name,
			"-",
		)
		applicationResults = append(applicationResults, *result)
	}

	return client.DeployResults{
		Message: reason,
		Success: false,
		Results: applicationResults,
	}
}

func printDeployResult(result []client.DeployResults, out io.Writer) error {
	var results []client.DeployResult
	for _, r := range result {
		results = append(results, r.Results...)
	}

	if len(results) == 0 {
		return errors.New("No deploys were made")
	}

	sort.Slice(results, func(i, j int) bool {
		nameA := results[i].DeploymentSpec.Name()
		nameB := results[j].DeploymentSpec.Name()
		return strings.Compare(nameA, nameB) < 1
	})

	header, rows := getDeployResultTable(results)
	if len(rows) == 0 {
		return nil
	}

	DefaultTablePrinter(header, rows, out)

	warningHeader, warningRows := getWarningTable(results)
	if len(warningRows) != 0 {
		fmt.Println("")
		fmt.Println("Some warnings where found:")
		DefaultTablePrinter(warningHeader, warningRows, out)
	}
	for _, deploy := range results {
		if !deploy.Success {
			return errors.New("One or more deploys failed")
		}
	}

	return nil
}

func getWarningTable(deploys []client.DeployResult) (string, []string) {
	var rows []string
	for _, item := range deploys {
		cluster := item.DeploymentSpec.Cluster()
		environment := item.DeploymentSpec.Environment()
		name := item.DeploymentSpec.Name()
		pattern := "%s\t%s\t%s\t%s\t%s"
		for _, warning := range item.Warnings {
			result := fmt.Sprintf(pattern, cluster, environment, name, item.DeployID, warning)
			rows = append(rows, result)
		}
	}

	header := "CLUSTER\tENVIRONMENT\tAPPLICATION\tDEPLOY_ID\tWARNING"
	return header, rows
}

func getDeployResultTable(deploys []client.DeployResult) (string, []string) {
	var rows []string
	for _, item := range deploys {
		if item.Ignored {
			continue
		}
		cluster := item.DeploymentSpec.Cluster()
		environment := item.DeploymentSpec.Environment()
		name := item.DeploymentSpec.Name()
		version := item.DeploymentSpec.Version()
		pattern := "%s\t%s\t%s\t%s\t%s\t%s\t%s"
		status := "\x1b[32mDeployed\x1b[0m"
		if !item.Success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, cluster, environment, name, version, item.DeployID, item.Reason)
		rows = append(rows, result)
	}

	header := "\x1b[00mSTATUS\x1b[0m\tCLUSTER\tENVIRONMENT\tAPPLICATION\tVERSION\tDEPLOY_ID\tMESSAGE"
	return header, rows
}

func updateVersion(apiClient client.AuroraConfigClient, applications []string, version string, out io.Writer) error {
	path := "/version"

	filenames, err := apiClient.GetFileNames()
	if err != nil {
		return err
	}
	fileName, err := filenames.Find(applications[0])
	if err != nil {
		return err
	}

	// Load config file
	auroraConfigFile, eTag, err := apiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	// Set new version
	if err := auroraconfig.SetValue(auroraConfigFile, path, version); err != nil {
		return err
	}

	// Save config file (Gobo)
	if _, err = DefaultAPIClient.UpdateAuroraConfigFile(auroraConfigFile, eTag); err != nil {
		return err
	}

	fmt.Fprintf(out, "%s has been updated with %s %s\n", fileName, path, version)

	return nil
}
