package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

type deploymentUnitID struct {
	envName, clusterName string
}

type deploymentUnit struct {
	id             *deploymentUnitID
	deploySpecList []client.DeploySpec
	cluster        *config.Cluster
	auroraConfig   string
	overrideToken  string
}

var (
	flagAffiliation string
	flagOverrides   []string
	flagNoPrompt    bool
	flagVersion     string
	flagCluster     string
	flagExcludes    []string
)

const deployLong = `Deploys applications from the current AuroraConfig.
For use in CI environments use --no-prompt to disable interactivity.
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

func newDeploymentUnitID(clusterName, envName string) *deploymentUnitID {
	return &deploymentUnitID{
		clusterName: clusterName,
		envName:     envName,
	}
}

func newDeploymentUnit(unitID *deploymentUnitID, deploySpecs []client.DeploySpec, cluster *config.Cluster, auroraConfig string, overrideToken string) *deploymentUnit {
	return &deploymentUnit{
		id:             unitID,
		deploySpecList: deploySpecs,
		cluster:        cluster,
		auroraConfig:   auroraConfig,
		overrideToken:  overrideToken,
	}
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&flagAffiliation, "auroraconfig", "a", "", "Overrides the logged in AuroraConfig")
	deployCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Limit deploy to given cluster name")
	deployCmd.Flags().BoolVarP(&flagNoPrompt, "no-prompt", "", false, "Suppress prompts")
	deployCmd.Flags().StringArrayVarP(&flagOverrides, "overrides", "o", []string{}, "Override in the form '[env/]file:{<json override>}'")
	deployCmd.Flags().StringArrayVarP(&flagExcludes, "exclude", "e", []string{}, "Select applications or environments to exclude from deploy")
	deployCmd.Flags().StringVarP(&flagVersion, "version", "v", "", "Set the given version in AuroraConfig before deploy")

	deployCmd.Flags().BoolVarP(&flagNoPrompt, "force", "f", false, "Suppress prompts")
	deployCmd.Flags().MarkHidden("force")
	deployCmd.Flags().StringVarP(&flagAffiliation, "affiliation", "", "", "Overrides the logged in affiliation")
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

	auroraConfig := AO.Affiliation
	if flagAffiliation != "" {
		auroraConfig = flagAffiliation
	}

	apiClient, err := getAPIClient(auroraConfig, pFlagToken, flagCluster)
	if err != nil {
		return err
	}

	applications, err := getApplications(apiClient, search, flagVersion, flagExcludes, cmd.OutOrStdout())
	if err != nil {
		return err
	} else if len(applications) == 0 {
		return errors.New("No applications to deploy")
	}

	filteredDeploymentSpecs, err := getFilteredDeploymentSpecs(apiClient, applications, flagCluster)
	if err != nil {
		return err
	}

	overrideConfig, err := parseOverride()
	if err != nil {
		return err
	}

	deploymentUnits, err := createDeploymentUnits(auroraConfig, pFlagToken, AO.Clusters, filteredDeploymentSpecs)
	if err != nil {
		return err
	}

	if !userConfirmation(filteredDeploymentSpecs, cmd.OutOrStdout()) {
		return errors.New("No applications to deploy")
	}

	result, err := deployToReachableClusters(getDeployClient, deploymentUnits, overrideConfig)
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

func createDeploymentUnits(auroraConfig, overrideToken string, clusters map[string]*config.Cluster, deploymentSpecs []client.DeploySpec) (map[deploymentUnitID]*deploymentUnit, error) {
	unitsMap := make(map[deploymentUnitID]*deploymentUnit)

	for _, spec := range deploymentSpecs {
		clusterName := spec.Value("cluster").(string)
		envName := spec.Value("envName").(string)

		unitID := newDeploymentUnitID(clusterName, envName)

		if _, exists := unitsMap[*unitID]; !exists {
			if _, exists := clusters[clusterName]; !exists {
				return nil, errors.New(fmt.Sprintf("No such cluster %s", clusterName))
			}
			cluster := clusters[clusterName]
			unit := newDeploymentUnit(unitID, []client.DeploySpec{}, cluster, auroraConfig, overrideToken)
			unitsMap[*unitID] = unit
		}

		unitsMap[*unitID].deploySpecList = append(unitsMap[*unitID].deploySpecList, spec)
	}

	return unitsMap, nil
}

func deployToReachableClusters(getClient func(unit *deploymentUnit) client.DeployClient, deploymentUnits map[deploymentUnitID]*deploymentUnit, overrideConfig map[string]string) ([]*client.DeployResults, error) {
	deployResult := make(chan *client.DeployResults)

	for _, unit := range deploymentUnits {
		go deployUnit(getClient(unit), unit, overrideConfig, deployResult)
	}

	var allResults []*client.DeployResults
	for i := 0; i < len(deploymentUnits); i++ {
		allResults = append(allResults, <-deployResult)
	}

	return allResults, nil
}

func deployUnit(deployClient client.DeployClient, unit *deploymentUnit, overrideConfig map[string]string, deployResults chan<- *client.DeployResults) {
	if !unit.cluster.Reachable {
		deployResults <- errorDeployResults("Cluster is not reachable", unit)
		return
	}

	var applicationList []string
	for _, spec := range unit.deploySpecList {
		applicationList = append(applicationList, spec.Value("applicationDeploymentRef").(string))
	}

	payload := client.NewDeployPayload(applicationList, overrideConfig)

	result, err := deployClient.Deploy(payload)
	if err != nil {
		deployResults <- errorDeployResults(err.Error(), unit)
	} else {
		deployResults <- result
	}
}

func errorDeployResults(reason string, unit *deploymentUnit) *client.DeployResults {
	var applicationResults []client.DeployResult

	for _, spec := range unit.deploySpecList {
		affiliation := spec.Value("affiliation").(string)
		applicationDeploymentRef := client.NewApplicationDeploymentRef(spec.Value("applicationDeploymentRef").(string))

		result := new(client.DeployResult)
		result.DeployId = "-"
		result.Ignored = false
		result.Success = false
		result.Reason = reason
		result.DeploymentSpec = make(client.DeploymentSpec)
		result.DeploymentSpec["cluster"] = client.NewAuroraConfigFieldSource(unit.cluster.Name)
		result.DeploymentSpec["name"] = client.NewAuroraConfigFieldSource(applicationDeploymentRef.Application)
		result.DeploymentSpec["version"] = client.NewAuroraConfigFieldSource("-")
		result.DeploymentSpec["envName"] = client.NewAuroraConfigFieldSource(affiliation + "-" + applicationDeploymentRef.Environment)

		applicationResults = append(applicationResults, *result)
	}

	return &client.DeployResults{
		Message: reason,
		Success: false,
		Results: applicationResults,
	}
}

func parseOverride() (map[string]string, error) {
	returnMap := make(map[string]string)
	for i := 0; i < len(flagOverrides); i++ {
		indexByte := strings.IndexByte(flagOverrides[i], ':')
		filename := flagOverrides[i][:indexByte]
		jsonOverride := flagOverrides[i][indexByte+1:]

		if !json.Valid([]byte(jsonOverride)) {
			msg := fmt.Sprintf("%s is not a valid json", jsonOverride)
			return nil, errors.New(msg)
		}

		returnMap[filename] = jsonOverride
	}
	return returnMap, nil
}

func userConfirmation(filteredDeploymentSpecs []client.DeploySpec, out io.Writer) bool {
	header, rows := GetDeploySpecTable(filteredDeploymentSpecs)
	DefaultTablePrinter(header, rows, out)

	var filteredApplications []string
	for _, spec := range filteredDeploymentSpecs {
		applicationDeploymentRef := spec.Value("applicationDeploymentRef").(string)
		filteredApplications = append(filteredApplications, applicationDeploymentRef)
	}

	shouldDeploy := true
	if !flagNoPrompt {
		defaultAnswer := len(filteredApplications) == 1
		message := fmt.Sprintf("Do you want to deploy %d application(s)?", len(filteredApplications))
		shouldDeploy = prompt.Confirm(message, defaultAnswer)
	}

	return shouldDeploy
}

func printDeployResult(result []*client.DeployResults, out io.Writer) error {
	var results []client.DeployResult
	for _, r := range result {
		results = append(results, r.Results...)
	}

	if len(results) == 0 {
		return errors.New("No deploys were made")
	}

	sort.Slice(results, func(i, j int) bool {
		resultI := results[i].DeploymentSpec.GetString("name")
		resultJ := results[j].DeploymentSpec.GetString("name")
		return strings.Compare(resultI, resultJ) < 1
	})

	header, rows := getDeployResultTable(results)
	if len(rows) == 0 {
		return nil
	}

	DefaultTablePrinter(header, rows, out)
	for _, deploy := range results {
		if !deploy.Success {
			return errors.New("One or more deploys failed")
		}
	}

	return nil
}

func getDeployResultTable(deploys []client.DeployResult) (string, []string) {
	var rows []string
	for _, item := range deploys {
		if item.Ignored {
			continue
		}
		cluster := item.DeploymentSpec.GetString("cluster")
		envName := item.DeploymentSpec.GetString("envName")
		name := item.DeploymentSpec.GetString("name")
		version := item.DeploymentSpec.GetString("version")
		pattern := "%s\t%s\t%s\t%s\t%s\t%s\t%s"
		status := "\x1b[32mDeployed\x1b[0m"
		if !item.Success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, cluster, envName, name, version, item.DeployId, item.Reason)
		rows = append(rows, result)
	}

	header := "\x1b[00mSTATUS\x1b[0m\tCLUSTER\tENVIRONMENT\tAPPLICATION\tVERSION\tDEPLOY_ID\tMESSAGE"
	return header, rows
}

func getAPIClient(auroraConfig, overrideToken, overrideCluster string) (*client.ApiClient, error) {
	api := DefaultApiClient
	api.Affiliation = auroraConfig

	if overrideCluster != "" && !AO.Localhost {
		c := AO.Clusters[overrideCluster]
		if !c.Reachable {
			return nil, errors.Errorf("%s cluster is not reachable", overrideCluster)
		}

		api.Host = c.BooberUrl
		api.Token = c.Token
		if overrideToken != "" {
			api.Token = overrideToken
		}
	}

	return api, nil
}

func getDeployClient(unit *deploymentUnit) client.DeployClient {
	var deployClient *client.ApiClient
	if AO.Localhost {
		deployClient = DefaultApiClient
		deployClient.Affiliation = unit.auroraConfig
	} else {
		token := unit.cluster.Token
		if unit.overrideToken != "" {
			token = unit.overrideToken
		}
		deployClient = client.NewApiClient(unit.cluster.BooberUrl, token, unit.auroraConfig, AO.RefName)
	}

	return deployClient
}
