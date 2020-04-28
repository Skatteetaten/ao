package cmd

import (
	"ao/pkg/client"
	"ao/pkg/deploymentspec"
	"github.com/pkg/errors"
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

func getDeployedApplications(getClient func(partition Partition) client.ApplicationDeploymentClient, deploySpecs []deploymentspec.DeploymentSpec, auroraConfigName, overrideToken string) ([]DeploymentInfo, error) {
	partitions, err := createDeploySpecPartitions(auroraConfigName, overrideToken, AO.Clusters, deploySpecs)
	if err != nil {
		return nil, err
	}

	partialResults, err := checkExistence(getClient, partitions)
	if err != nil {
		return nil, err
	}

	var allResults []DeploymentInfo

	for _, partialResult := range partialResults {
		if !partialResult.existsResults.Success {
			return nil, errors.New("Failed to retrieve application deployment information from cluster")
		}

		for _, existsResult := range partialResult.existsResults.Results {
			if existsResult.Exists {
				info := newDeploymentInfo(existsResult.ApplicationRef.Namespace, existsResult.ApplicationRef.Name, partialResult.partition.Cluster.Name)
				allResults = append(allResults, *info)
			}
		}
	}

	return allResults, nil
}

func checkExistence(getClient func(partition Partition) client.ApplicationDeploymentClient, partitions []DeploySpecPartition) ([]partialExistsResult, error) {
	partialResults := make(chan partialExistsResult)
	existsErrors := make(chan error)

	for _, partition := range partitions {
		go performExists(getClient(partition.Partition), partition, partialResults, existsErrors)
	}

	var allResults []partialExistsResult

	for i := 0; i < len(partitions); i++ {
		select {
		case err := <-existsErrors:
			return nil, err
		case result := <-partialResults:
			allResults = append(allResults, result)
		}
	}

	return allResults, nil
}

func performExists(deployClient client.ApplicationDeploymentClient, partition DeploySpecPartition, partialResult chan<- partialExistsResult, existsErrors chan<- error) {
	if !partition.Cluster.Reachable {
		existsErrors <- errors.New("Cluster is not reachable")
		return
	}

	var applicationList []string
	for _, spec := range partition.DeploySpecs {
		applicationList = append(applicationList, spec.GetString("applicationDeploymentRef"))
	}

	results, err := deployClient.Exists(client.NewExistsPayload(applicationList))

	if err != nil {
		existsErrors <- errors.Wrap(err, "Unable to determine wether applications exists on OpenShift or not")
	} else {
		partialResult <- newPartialExistsResults(partition, *results)
	}
}
