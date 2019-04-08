package cmd

import (
	"fmt"
	"io"
	"regexp"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/fuzzy"
)

type requestPartitionID struct {
	envName, clusterName string
}

type requestPartition struct {
	id               *requestPartitionID
	deploySpecList   []client.DeploySpec
	cluster          *config.Cluster
	auroraConfigName string
	overrideToken    string
}

var (
	flagAuroraConfig string
	flagOverrides    []string
	flagNoPrompt     bool
	flagVersion      string
	flagCluster      string
	flagExcludes     []string
)

func newRequestPartitionID(clusterName, envName string) *requestPartitionID {
	return &requestPartitionID{
		clusterName: clusterName,
		envName:     envName,
	}
}

func newRequestPartition(unitID *requestPartitionID, deploySpecs []client.DeploySpec, cluster *config.Cluster, auroraConfig string, overrideToken string) *requestPartition {
	return &requestPartition{
		id:               unitID,
		deploySpecList:   deploySpecs,
		cluster:          cluster,
		auroraConfigName: auroraConfig,
		overrideToken:    overrideToken,
	}
}

func getApplications(apiClient client.AuroraConfigClient, search, version string, excludes []string, out io.Writer) ([]string, error) {
	files, err := apiClient.GetFileNames()
	if err != nil {
		return nil, err
	}

	possibleDeploys := files.GetApplicationDeploymentRefs()
	applications := fuzzy.SearchForApplications(search, possibleDeploys)

	applications, err = filterExcludes(excludes, applications)
	if err != nil {
		return nil, err
	}

	if version != "" {
		if len(applications) > 1 {
			return nil, errors.New("Deploy with version does only support one application")
		}

		fileName, err := files.Find(applications[0])
		if err != nil {
			return nil, err
		}

		err = updateVersion(apiClient, version, fileName, out)
		if err != nil {
			return nil, err
		}
	}

	return applications, nil
}

func updateVersion(apiClient client.AuroraConfigClient, version, fileName string, out io.Writer) error {
	path, value := "/version", version

	fileName, err := SetValue(apiClient, fileName, path, value)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s has been updated with %s %s\n", fileName, path, value)

	return nil
}

func getFilteredDeploymentSpecs(apiClient client.DeploySpecClient, applications []string, overrideCluster string) ([]client.DeploySpec, error) {
	deploySpecs, err := apiClient.GetAuroraDeploySpec(applications, true)
	if err != nil {
		return nil, err
	}
	var filteredDeploymentSpecs []client.DeploySpec
	if overrideCluster != "" {
		for _, spec := range deploySpecs {
			if spec.Value("/cluster").(string) == overrideCluster {
				filteredDeploymentSpecs = append(filteredDeploymentSpecs, spec)
			}
		}
	} else {
		filteredDeploymentSpecs = deploySpecs
	}

	return filteredDeploymentSpecs, nil
}

func filterExcludes(expressions, applications []string) ([]string, error) {
	apps := make([]string, len(applications))
	copy(apps, applications)
	for _, expr := range expressions {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		tmp := apps[:0]
		for _, app := range apps {
			match := r.MatchString(app)
			if !match {
				tmp = append(tmp, app)
			}
		}
		apps = tmp
	}

	return apps, nil
}

func createRequestPartitions(auroraConfig, overrideToken string, clusters map[string]*config.Cluster, deploymentSpecs []client.DeploySpec) (map[requestPartitionID]*requestPartition, error) {
	partitionMap := make(map[requestPartitionID]*requestPartition)

	for _, spec := range deploymentSpecs {
		clusterName := spec.Value("cluster").(string)
		envName := spec.Value("envName").(string)

		partitionID := newRequestPartitionID(clusterName, envName)

		if _, exists := partitionMap[*partitionID]; !exists {
			if _, exists := clusters[clusterName]; !exists {
				return nil, errors.New(fmt.Sprintf("No such cluster %s", clusterName))
			}
			cluster := clusters[clusterName]
			partition := newRequestPartition(partitionID, []client.DeploySpec{}, cluster, auroraConfig, overrideToken)
			partitionMap[*partitionID] = partition
		}

		partitionMap[*partitionID].deploySpecList = append(partitionMap[*partitionID].deploySpecList, spec)
	}

	return partitionMap, nil
}

func getApplicationDeploymentClient(partition *requestPartition) client.ApplicationDeploymentClient {
	var cli *client.ApiClient
	if AO.Localhost {
		cli = DefaultApiClient
		cli.Affiliation = partition.auroraConfigName
	} else {
		token := partition.cluster.Token
		if partition.overrideToken != "" {
			token = partition.overrideToken
		}
		cli = client.NewApiClient(partition.cluster.BooberUrl, token, partition.auroraConfigName, AO.RefName)
	}

	return cli
}
