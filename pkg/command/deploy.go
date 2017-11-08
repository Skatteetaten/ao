package command

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"sort"
	"strings"
)

type DeployOptions struct {
	Affiliation   string
	Token         string
	Cluster       string
	Version       string
	Overrides     []string
	DeployApiOnly bool
	DeployOnce    bool
	DeployAll     bool
	Force         bool
}

func Deploy(args []string, api *client.ApiClient, clusters map[string]*config.Cluster, options *DeployOptions) []client.DeployResult {

	overrides, err := parseOverride(options.Overrides)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	api.Affiliation = options.Affiliation

	if options.Cluster != "" {
		cluster := clusters[options.Cluster]
		if cluster == nil {
			fmt.Println("No such cluster", options.Cluster)
			return nil
		}

		api.Host = cluster.BooberUrl
		api.Token = cluster.Token
		if options.Token != "" {
			api.Token = options.Token
		}
	}

	files, err := api.GetFileNames()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	possibleDeploys := files.FilterDeployments()
	appsToDeploy := []string{}
	if options.DeployAll {
		args = []string{}
		appsToDeploy = possibleDeploys
	}

	for _, arg := range args {
		applications, _ := fuzzy.SearchForApplications(arg, possibleDeploys)
		if !options.Force && len(applications) > 1 {
			selectedApps := applications
			printDeployments(applications)
			deployAll := prompt.ConfirmDeployAll(applications)
			if !deployAll {
				selectedApps = prompt.MultiSelectDeployments(applications)
			}
			appsToDeploy = append(appsToDeploy, selectedApps...)
		} else {
			appsToDeploy = append(appsToDeploy, applications...)
		}
	}

	if len(appsToDeploy) == 0 {
		fmt.Println("No applications to deploy")
		return nil
	}

	if !options.Force {
		printDeployments(appsToDeploy)
		shouldDeploy := prompt.ConfirmDeploy(appsToDeploy)
		if !shouldDeploy {
			return nil
		}
	}

	payload, err := client.NewDeployPayload(appsToDeploy, overrides)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if options.Version != "" {
		if len(appsToDeploy) > 1 {
			fmt.Println("Deploy with version does only support one application")
			return nil
		}
		operation := client.JsonPatchOp{
			OP:    "add",
			Path:  "/version",
			Value: options.Version,
		}

		fileName := appsToDeploy[0] + ".json"
		err := api.PatchAuroraConfigFile(fileName, operation)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}

	if options.DeployOnce {
		result, err := api.Deploy(payload)
		if err != nil {
			fmt.Println(err)
		}
		PrintDeployResults(result)
		return nil
	}

	allResults := deployToReachableClusters(options.Affiliation, options.Token, clusters, payload)

	return allResults
}

func deployToReachableClusters(affiliation, token string, clusters map[string]*config.Cluster, payload *client.DeployPayload) []client.DeployResult {

	reachableClusters := 0
	deployResult := make(chan []client.DeployResult)
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

	var allResults []client.DeployResult
	for i := 0; i < reachableClusters; i++ {
		select {
		case err := <-deployErrors:
			fmt.Println(err)
		case result := <-deployResult:
			allResults = append(allResults, result...)
		}
	}

	return allResults
}

func parseOverride(override []string) (returnMap map[string]json.RawMessage, err error) {
	returnMap = make(map[string]json.RawMessage)

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
	return returnMap, err
}

func printDeployments(deployments []string) {
	sort.Strings(deployments)
	lines := GetDeploymentTable(deployments)
	DefaultTablePrinter(lines)
}
