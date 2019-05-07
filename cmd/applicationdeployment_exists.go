package cmd

import (
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
)

type partialExistsResult struct {
	partition     DeploySpecPartition
	existsResults client.ExistsResults
}

func newPartialExistsResults(partition DeploySpecPartition, existsResults client.ExistsResults) partialExistsResult {
	return partialExistsResult{
		partition:     partition,
		existsResults: existsResults,
	}
}

func getDeployedApplications(deploySpecs []client.DeploySpec, auroraConfigName, overrideToken string) ([]DeploymentInfo, error) {
	partitions, err := createDeploySpecPartitions(auroraConfigName, overrideToken, AO.Clusters, deploySpecs)
	if err != nil {
		return nil, err
	}

	partialResults := make(chan partialExistsResult)

	for _, partition := range partitions {
		go performExists(getApplicationDeploymentClient(partition.Partition), partition, partialResults)
	}

	var allResults []DeploymentInfo

	for i := 0; i < len(partitions); i++ {
		results := <-partialResults
		if !results.existsResults.Success {
			return nil, errors.New("Failed to retrieve application deployment information from cluster")
		}

		for _, existsResult := range results.existsResults.Results {
			if existsResult.Exists {
				info := newDeploymentInfo(existsResult.ApplicationRef.Namespace, existsResult.ApplicationRef.Name, results.partition.Cluster.Name)
				allResults = append(allResults, *info)
			}
		}
	}

	return allResults, nil
}

func performExists(deployClient client.ApplicationDeploymentClient, partition DeploySpecPartition, partialResult chan<- partialExistsResult) {
	if !partition.Cluster.Reachable {
		partialResult <- errorExistsResults("Cluster is not reachable", partition)
		return
	}

	var applicationList []string
	for _, spec := range partition.DeploySpecs {
		applicationList = append(applicationList, spec.Value("applicationDeploymentRef").(string))
	}

	results, err := deployClient.Exists(client.NewExistsPayload(applicationList))

	if err != nil {
		partialResult <- errorExistsResults(err.Error(), partition)
	} else {
		partialResult <- newPartialExistsResults(partition, *results)
	}
}

func errorExistsResults(reason string, partition DeploySpecPartition) partialExistsResult {
	existsResults := &client.ExistsResults{
		Message: reason,
		Success: false,
		Results: nil,
	}

	return newPartialExistsResults(partition, *existsResults)
}
