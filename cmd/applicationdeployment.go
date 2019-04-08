package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

var (
	applicationDeploymentCmd = &cobra.Command{
		Use:         "applicationdeployment",
		Short:       "Delete Aurora resources",
		Annotations: map[string]string{"type": "remote"},
	}

	applicationDeploymentDeleteCmd = &cobra.Command{
		Use:   "delete <applicationDeploymentRef>",
		Short: "Delete application deployment with the given reference",
		RunE:  deleteApplicationDeployment,
	}
)

type partitionDeleteResult struct {
	partition     requestPartition
	deleteResults client.DeleteResults
}

type deleteSummary struct {
	cluster string
	env     string
	name    string
	success bool
	reason  string
}

func newPartitionDeleteResult(partition requestPartition, deleteResults client.DeleteResults) *partitionDeleteResult {
	return &partitionDeleteResult{
		partition:     partition,
		deleteResults: deleteResults,
	}
}

func newPrintDeleteResult(cluster, env, name, reason string, success bool) *deleteSummary {
	return &deleteSummary{
		cluster: cluster,
		env:     env,
		name:    name,
		success: success,
		reason:  reason,
	}
}

func init() {
	RootCmd.AddCommand(applicationDeploymentCmd)
	applicationDeploymentCmd.AddCommand(applicationDeploymentDeleteCmd)
	applicationDeploymentDeleteCmd.Flags().StringVarP(&flagAuroraConfig, "auroraconfig", "a", "", "Overrides the logged in AuroraConfig")
	applicationDeploymentDeleteCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Limit deletion to given cluster name")
	applicationDeploymentDeleteCmd.Flags().BoolVarP(&flagNoPrompt, "no-prompt", "", false, "Suppress prompts")
	applicationDeploymentDeleteCmd.Flags().StringArrayVarP(&flagExcludes, "exclude", "e", []string{}, "Select applications or environments to exclude from deletion")

	applicationDeploymentDeleteCmd.Flags().BoolVarP(&flagNoPrompt, "force", "f", false, "Suppress prompts")
	applicationDeploymentDeleteCmd.Flags().MarkHidden("force")
	applicationDeploymentDeleteCmd.Flags().StringVarP(&flagAuroraConfig, "affiliation", "", "", "Overrides the logged in affiliation")
	applicationDeploymentDeleteCmd.Flags().MarkHidden("affiliation")
}

func deleteApplicationDeployment(cmd *cobra.Command, args []string) error {

	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	err := validateDeleteParams()
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

	applications, err := getApplications(apiClient, search, "", flagExcludes, cmd.OutOrStdout())
	if err != nil {
		return err
	} else if len(applications) == 0 {
		return errors.New("No applications to deploy")
	}

	filteredDeploymentSpecs, err := getFilteredDeploymentSpecs(apiClient, applications, flagCluster)
	if err != nil {
		return err
	}

	partitions, err := createRequestPartitions(auroraConfigName, pFlagToken, AO.Clusters, filteredDeploymentSpecs)
	if err != nil {
		return err
	}

	if !deleteConfirmation(flagNoPrompt, filteredDeploymentSpecs, cmd.OutOrStdout()) {
		return errors.New("No applications to delete")
	}

	result, err := deleteFromReachableClusters(getApplicationDeploymentClient, partitions)
	if err != nil {
		return err
	}

	printDeleteResult(result, cmd.OutOrStdout())

	return nil
}

func validateDeleteParams() error {

	if flagCluster != "" {
		if _, exists := AO.Clusters[flagCluster]; !exists {
			return errors.New(fmt.Sprintf("No such cluster %s", flagCluster))
		}
	}

	return nil
}

func deleteConfirmation(force bool, filteredDeploymentSpecs []client.DeploySpec, out io.Writer) bool {
	header, rows := GetCompactDeploySpecTable(filteredDeploymentSpecs)
	DefaultTablePrinter(header, rows, out)

	shouldDeploy := true
	if !force {
		defaultAnswer := len(rows) == 1
		message := fmt.Sprintf("Do you want to delete %d application(s)?", len(rows))
		shouldDeploy = prompt.Confirm(message, defaultAnswer)
	}

	return shouldDeploy
}

func deleteFromReachableClusters(getClient func(partition *requestPartition) client.ApplicationDeploymentClient, partitions map[requestPartitionID]*requestPartition) ([]*partitionDeleteResult, error) {
	partitionResult := make(chan *partitionDeleteResult)

	for _, partition := range partitions {
		go performDelete(getClient(partition), *partition, partitionResult)
	}

	var allResults []*partitionDeleteResult
	for i := 0; i < len(partitions); i++ {
		allResults = append(allResults, <-partitionResult)
	}

	return allResults, nil
}

func performDelete(deployClient client.ApplicationDeploymentClient, partition requestPartition, partitionResult chan<- *partitionDeleteResult) {
	if !partition.cluster.Reachable {
		partitionResult <- errorDeleteResults("Cluster is not reachable", partition)
		return
	}

	var applicationList []string
	for _, spec := range partition.deploySpecList {
		applicationList = append(applicationList, spec.Value("applicationDeploymentRef").(string))
	}

	results, err := deployClient.Delete(client.NewDeletePayload(applicationList))

	if err != nil {
		partitionResult <- errorDeleteResults(err.Error(), partition)
	} else {
		partitionResult <- newPartitionDeleteResult(partition, *results)
	}
}

func errorDeleteResults(reason string, partition requestPartition) *partitionDeleteResult {
	var deleteResultList []client.DeleteResult

	for _, spec := range partition.deploySpecList {
		result := new(client.DeleteResult)
		result.Success = false
		result.Reason = reason
		result.ApplicationDeploymentRef = *client.NewApplicationDeploymentRef(spec.Value("applicationDeploymentRef").(string))

		deleteResultList = append(deleteResultList, *result)
	}

	deleteResults := &client.DeleteResults{
		Message: reason,
		Success: false,
		Results: deleteResultList,
	}

	return newPartitionDeleteResult(partition, *deleteResults)
}

func printDeleteResult(allResults []*partitionDeleteResult, out io.Writer) error {
	var printSummary []deleteSummary

	for _, partitionResult := range allResults {
		for _, deleteResult := range partitionResult.deleteResults.Results {
			cluster := partitionResult.partition.cluster.Name
			env := deleteResult.ApplicationDeploymentRef.Environment
			name := deleteResult.ApplicationDeploymentRef.Application
			success := deleteResult.Success
			reason := deleteResult.Reason

			printSummary = append(printSummary, *newPrintDeleteResult(cluster, env, name, reason, success))
		}
	}

	if len(printSummary) == 0 {
		return errors.New("No deploys were made")
	}

	sort.Slice(printSummary, func(i, j int) bool {
		nameA := printSummary[i].name
		nameB := printSummary[j].name
		return strings.Compare(nameA, nameB) < 1
	})

	header, rows := getDeleteResultTableContent(printSummary)
	if len(rows) == 0 {
		return nil
	}

	DefaultTablePrinter(header, rows, out)
	for _, delete := range printSummary {
		if !delete.success {
			return errors.New("One or more deploys failed")
		}
	}

	return nil
}

func getDeleteResultTableContent(printSummary []deleteSummary) (string, []string) {
	var rows []string
	for _, item := range printSummary {
		pattern := "%s\t%s\t%s\t%s\t%s"
		status := "\x1b[32mDeleted\x1b[0m"
		if !item.success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, item.cluster, item.env, item.name, item.reason)
		rows = append(rows, result)
	}

	header := "\x1b[00mSTATUS\x1b[0m\tCLUSTER\tENVIRONMENT\tAPPLICATION\tMESSAGE"
	return header, rows
}
