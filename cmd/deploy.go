package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

var (
	flagAffiliation string
	flagOverrides   []string
	flagNoPrompt    bool
	flagVersion     string
	flagCluster     string
)

const deployLong = `When specifying applicationId you can use fuzzy matching. If there are multiple applications to deploy,
you can choose to deploy all or select desired applications.
For use in CI environments please use --noprompt and --cluster flag or else you may get unexpected results.
`

const exampleDeploy = `Given the following AuroraConfig:
  - about.json
  - foobar.json
  - bar.json
  - foo/about.json
  - foo/bar.json
  - foo/foobar.json

Fuzzy matching
  ao deploy fo/ba == foo/bar and foo/foobar

Exact matching
  ao deploy foo/bar == only foo/bar

Deploy an application with override for application file
  ao deploy foo/bar -o 'foo/bar.json:{"pause": true}'
`

var deployCmd = &cobra.Command{
	Aliases:     []string{"setup", "apply"},
	Use:         "deploy <applicationId>",
	Short:       "Deploy one or more ApplicationId (environment/application) to one or more clusters",
	Long:        deployLong,
	Example:     exampleDeploy,
	Annotations: map[string]string{"type": "actions"},
	RunE:        deploy,
}

func init() {
	RootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&flagAffiliation, "affiliation", "", "", "Overrides the logged in affiliation")
	deployCmd.Flags().MarkHidden("affiliation")
	deployCmd.Flags().StringVarP(&flagAffiliation, "auroraconfig", "a", "", "Overrides the logged in AuroraConfig")
	deployCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Limit deploy to given cluster name")
	deployCmd.Flags().BoolVarP(&flagNoPrompt, "force", "f", false, "Suppress prompts")
	deployCmd.Flags().MarkHidden("force")
	deployCmd.Flags().BoolVarP(&flagNoPrompt, "noprompt", "", false, "Suppress prompts")

	deployCmd.Flags().StringArrayVarP(&flagOverrides, "overrides", "o", []string{}, "Override in the form '[env/]file:{<json override>}'")
	deployCmd.Flags().StringVarP(&flagVersion, "version", "v", "", "Set the given version in AuroraConfig before deploy")
}

func deploy(cmd *cobra.Command, args []string) error {

	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	overrides, err := parseOverride(flagOverrides)
	if err != nil {
		return err
	}

	if flagAffiliation == "" {
		flagAffiliation = AO.Affiliation
	}

	api := DefaultApiClient
	api.Affiliation = flagAffiliation

	if flagCluster != "" {
		c := AO.Clusters[flagCluster]
		if c == nil {
			return errors.New("No such cluster " + flagCluster)
		}

		api.Host = c.BooberUrl
		api.Token = c.Token
		if pFlagToken != "" {
			api.Token = pFlagToken
		}
	}

	files, err := api.GetFileNames()
	if err != nil {
		return err
	}

	possibleDeploys := files.GetApplicationIds()
	applications := fuzzy.SearchForApplications(search, possibleDeploys)

	if len(applications) == 0 {
		return errors.New("No applications to deploy")
	}

	header, rows := GetApplicationIdTable(applications)
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())

	shouldDeploy := true
	if !flagNoPrompt {
		message := fmt.Sprintf("Do you want to deploy %d application(s)?", len(applications))
		shouldDeploy = prompt.Confirm(message)
	}

	if !flagNoPrompt && !shouldDeploy && len(applications) > 1 {
		applications = prompt.MultiSelect("Which applications do you want to deploy?", applications)
		shouldDeploy = len(applications) > 0
	}

	if !shouldDeploy {
		return errors.New("No applications to deploy")
	}

	if flagVersion != "" {
		if len(applications) > 1 {
			return errors.New("Deploy with version does only support one application")
		}
		operation := client.JsonPatchOp{
			OP:    "add",
			Path:  "/version",
			Value: flagVersion,
		}

		fileName := applications[0] + ".json"
		err := api.PatchAuroraConfigFile(fileName, operation)
		if err != nil {
			return err
		}
	}

	payload := client.NewDeployPayload(applications, overrides)

	var result []*client.DeployResults
	if AO.Localhost || flagCluster != "" {
		res, err := api.Deploy(payload)
		if err != nil {
			return err
		}
		result = append(result, res)
	} else {
		result = deployToReachableClusters(flagAffiliation, pFlagToken, AO.Clusters, payload)
	}

	var results []client.DeployResult
	for _, r := range result {
		if !r.Success {
			cmd.Println("deploy error:", r.Message)
		}
		results = append(results, r.Results...)
	}

	if len(results) == 0 {
		return errors.New("No deploys were made")
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.Compare(results[i].ADS.Name, results[j].ADS.Name) < 1
	})

	header, rows = getDeployResultTable(results)
	DefaultTablePrinter(header, rows, cmd.OutOrStdout())
	return nil
}

func deployToReachableClusters(affiliation, token string, clusters map[string]*config.Cluster, payload *client.DeployPayload) []*client.DeployResults {

	reachableClusters := 0
	deployResult := make(chan *client.DeployResults)
	deployErrors := make(chan error)
	for _, c := range clusters {
		if !c.Reachable {
			continue
		}
		reachableClusters++

		clusterToken := c.Token
		if token != "" {
			clusterToken = token
		}

		cli := client.NewApiClient(c.BooberUrl, clusterToken, affiliation)

		go func() {
			result, err := cli.Deploy(payload)
			if err != nil {
				deployErrors <- err
			} else {
				deployResult <- result
			}
		}()
	}

	var allResults []*client.DeployResults
	for i := 0; i < reachableClusters; i++ {
		select {
		case err := <-deployErrors:
			fmt.Println(err)
		case result := <-deployResult:
			allResults = append(allResults, result)
		}
	}

	return allResults
}

func parseOverride(override []string) (map[string]json.RawMessage, error) {
	returnMap := make(map[string]json.RawMessage)
	for i := 0; i < len(override); i++ {
		indexByte := strings.IndexByte(override[i], ':')
		filename := override[i][:indexByte]
		jsonOverride := override[i][indexByte+1:]

		if !json.Valid([]byte(jsonOverride)) {
			msg := fmt.Sprintf("%s is not a valid json", jsonOverride)
			return nil, errors.New(msg)
		}

		returnMap[filename] = json.RawMessage(jsonOverride)
	}
	return returnMap, nil
}

func getDeployResultTable(deploys []client.DeployResult) (string, []string) {
	var rows []string
	for _, item := range deploys {
		ads := item.ADS
		pattern := "%s\t%s\t%s\t%s\t%s"
		status := "\x1b[32mDeployed\x1b[0m"
		if !item.Success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, ads.Name, ads.Namespace, ads.Cluster, item.DeployId)
		rows = append(rows, result)
	}

	header := "\x1b[00mSTATUS\x1b[0m\tAPPLICATION\tENVIRONMENT\tCLUSTER\tDEPLOY_ID"
	return header, rows
}
