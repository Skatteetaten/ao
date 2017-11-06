package command

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
)

type DeployOptions struct {
	Affiliation string
	Token       string
	Cluster     string
	Overrides   []string
	DeployOnce  bool
	DeployAll   bool
	Force       bool
}

func Deploy(args []string, api *client.ApiClient, clusters map[string]*config.Cluster, options DeployOptions) {

	overrides, err := ParseOverride(options.Overrides)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Override must start and end with ' or else escape \" ")
		return
	}

	api.Affiliation = options.Affiliation

	if options.Cluster != "" {
		cluster := clusters[options.Cluster]
		if cluster == nil {
			fmt.Println("No such cluster", options.Cluster)
			return
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
		return
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
			PrintDeployments(applications)
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
		return
	}

	if !options.Force {
		PrintDeployments(appsToDeploy)
		shouldDeploy := prompt.ConfirmDeploy(appsToDeploy)
		if !shouldDeploy {
			return
		}
	}

	if options.DeployOnce {
		result, err := api.Deploy(appsToDeploy, overrides)
		if err != nil {
			fmt.Println(err)
		}
		PrintDeployResults(result)
		return
	}

	reachableClusters := 0
	deployResult := make(chan []client.DeployResult)
	deployErrors := make(chan error)
	for _, c := range clusters {
		if !c.Reachable {
			continue
		}
		reachableClusters++

		token := c.Token
		if options.Token != "" {
			token = options.Token
		}

		cli := client.NewApiClient(c.BooberUrl, token, options.Affiliation)

		go func() {
			result, err := cli.Deploy(appsToDeploy, overrides)
			if err != nil {
				deployErrors <- err
			} else {
				deployResult <- result
			}
		}()
	}

	allResults := []client.DeployResult{}
	counter := 0
	for {
		select {
		case err := <-deployErrors:
			fmt.Println(err)
			counter++
		case result := <-deployResult:
			allResults = append(allResults, result...)
			counter++
		}
		if counter == reachableClusters {
			break
		}
	}

	PrintDeployResults(allResults)
}
