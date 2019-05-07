package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/skatteetaten/ao/pkg/service"
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

type DeploymentPartition struct {
	Partition
	DeploymentInfos []DeploymentInfo
}

type deploymentPartitionID struct {
	namespace, clusterName string
}

type partialDeleteResult struct {
	partition     DeploymentPartition
	deleteResults client.DeleteResults
}

type deleteSummary struct {
	cluster string
	env     string
	name    string
	success bool
	reason  string
}

func newDeploymentPartition(deploymentInfos []DeploymentInfo, cluster config.Cluster, auroraConfig string, overrideToken string) *DeploymentPartition {
	return &DeploymentPartition{
		DeploymentInfos: deploymentInfos,
		Partition: Partition{
			Cluster:          cluster,
			AuroraConfigName: auroraConfig,
			OverrideToken:    overrideToken,
		},
	}
}

func newPartialDeleteResults(partition DeploymentPartition, deleteResults client.DeleteResults) partialDeleteResult {
	return partialDeleteResult{
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

	applications, err := service.GetApplications(apiClient, search, "", flagExcludes, cmd.OutOrStdout())
	if err != nil {
		return err
	} else if len(applications) == 0 {
		return errors.New("No applications to delete")
	}

	filteredDeploymentSpecs, err := service.GetFilteredDeploymentSpecs(apiClient, applications, flagCluster)
	if err != nil {
		return err
	}

	deployInfos, err := getDeployedApplications(filteredDeploymentSpecs, auroraConfigName, pFlagToken)
	if err != nil {
		return err
	} else if len(deployInfos) == 0 {
		return errors.New("No applications to delete")
	}

	partitions, err := createDeploymentPartitions(auroraConfigName, pFlagToken, AO.Clusters, deployInfos)
	if err != nil {
		return err
	}

	if !getDeleteConfirmation(flagNoPrompt, deployInfos, cmd.OutOrStdout()) {
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

func createDeploymentPartitions(auroraConfig, overrideToken string, clusters map[string]*config.Cluster, deployInfos []DeploymentInfo) ([]DeploymentPartition, error) {
	partitionMap := make(map[deploymentPartitionID]*DeploymentPartition)

	for _, info := range deployInfos {
		clusterName := info.ClusterName
		namespace := info.Namespace

		partitionID := deploymentPartitionID{clusterName, namespace}

		if _, exists := partitionMap[partitionID]; !exists {
			if _, exists := clusters[clusterName]; !exists {
				return nil, errors.New(fmt.Sprintf("No such cluster %s", clusterName))
			}
			cluster := clusters[clusterName]
			partition := newDeploymentPartition([]DeploymentInfo{}, *cluster, auroraConfig, overrideToken)
			partitionMap[partitionID] = partition
		}

		partitionMap[partitionID].DeploymentInfos = append(partitionMap[partitionID].DeploymentInfos, info)
	}

	partitions := make([]DeploymentPartition, len(partitionMap))

	idx := 0
	for _, partition := range partitionMap {
		partitions[idx] = *partition
		idx++
	}

	return partitions, nil
}

func deleteFromReachableClusters(getClient func(partition Partition) client.ApplicationDeploymentClient, partitions []DeploymentPartition) ([]partialDeleteResult, error) {
	partitionResult := make(chan partialDeleteResult)

	for _, partition := range partitions {
		go performDelete(getClient(partition.Partition), partition, partitionResult)
	}

	var allResults []partialDeleteResult
	for i := 0; i < len(partitions); i++ {
		allResults = append(allResults, <-partitionResult)
	}

	return allResults, nil
}

func performDelete(deployClient client.ApplicationDeploymentClient, partition DeploymentPartition, partitionResult chan<- partialDeleteResult) {
	if !partition.Cluster.Reachable {
		partitionResult <- errorDeleteResults("Cluster is not reachable", partition)
		return
	}

	var applicationRefs []client.ApplicationRef
	for _, info := range partition.DeploymentInfos {
		applicationRefs = append(applicationRefs, *client.NewApplicationRef(info.Namespace, info.Name))
	}

	results, err := deployClient.Delete(client.NewDeletePayload(applicationRefs))

	if err != nil {
		partitionResult <- errorDeleteResults(err.Error(), partition)
	} else {
		partitionResult <- newPartialDeleteResults(partition, *results)
	}
}

func getDeleteConfirmation(force bool, deployInfos []DeploymentInfo, out io.Writer) bool {
	header, rows := getDeployInfoTable(deployInfos)
	DefaultTablePrinter(header, rows, out)

	shouldDeploy := true
	if !force {
		defaultAnswer := len(rows) == 1
		message := fmt.Sprintf("Do you want to delete %d application(s)?", len(rows))
		shouldDeploy = prompt.Confirm(message, defaultAnswer)
	}

	return shouldDeploy
}

func getDeployInfoTable(infos []DeploymentInfo) (string, []string) {
	var rows []string
	header := "CLUSTER\tNAMESPACE\tAPPLICATION"
	pattern := "%v\t%v\t%v"
	sort.Slice(infos, func(i, j int) bool {
		return strings.Compare(infos[i].Name, infos[j].Name) != 1
	})
	for _, info := range infos {
		row := fmt.Sprintf(
			pattern,
			info.ClusterName,
			info.Namespace,
			info.Name,
		)
		rows = append(rows, row)
	}
	return header, rows
}

func errorDeleteResults(reason string, partition DeploymentPartition) partialDeleteResult {
	var results []client.DeleteResult

	for _, info := range partition.DeploymentInfos {
		result := new(client.DeleteResult)
		result.Success = false
		result.Reason = reason
		result.ApplicationRef = *client.NewApplicationRef(info.Namespace, info.Name)

		results = append(results, *result)
	}

	deleteResults := &client.DeleteResults{
		Message: reason,
		Success: false,
		Results: results,
	}

	return newPartialDeleteResults(partition, *deleteResults)
}

func printDeleteResult(allResults []partialDeleteResult, out io.Writer) error {
	var printSummary []deleteSummary

	for _, partitionResult := range allResults {
		for _, deleteResult := range partitionResult.deleteResults.Results {
			cluster := partitionResult.partition.Cluster.Name
			env := deleteResult.ApplicationRef.Namespace
			name := deleteResult.ApplicationRef.Name
			success := deleteResult.Success
			reason := deleteResult.Reason

			printSummary = append(printSummary, *newPrintDeleteResult(cluster, env, name, reason, success))
		}
	}

	if len(printSummary) == 0 {
		return errors.New("No applications were deleted")
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
