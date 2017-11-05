package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/tabwriter"
	"os"
	"sort"
)

type ApplicationId struct {
	Environment string `json:"environment"`
	Application string `json:"application"`
}

type deployResult struct {
	DeployId string `json:"deployId"`
	ADS      struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Cluster   string `json:"cluster"`
	} `json:"auroraDeploymentSpec"`
	Success bool `json:"success"`
}

type applyPayload struct {
	ApplicationIds []ApplicationId            `json:"applicationIds"`
	Overrides      map[string]json.RawMessage `json:"overrides"`
	Deploy         bool                       `json:"deploy"`
}

func (api *ApiClient) Deploy(applications []string, overrides map[string]json.RawMessage) error {

	applicationIds := createApplicationIds(applications)
	applyPayload := &applyPayload{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
		Deploy:         true,
	}

	payload, err := json.Marshal(applyPayload)
	if err != nil {
		fmt.Println("Failed to marshal ApplyPayload")
		return nil
	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)
	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	var deploys []deployResult
	err = json.Unmarshal(response.Items, &deploys)
	if err != nil {
		return err
	}

	sort.Slice(deploys, func(i, j int) bool {
		return strings.Compare(deploys[i].ADS.Name, deploys[j].ADS.Name) < 1
	})

	results := []string{"\x1b[00mSTATUS\x1b[0m\tAPPLICATION\tENVIRONMENT\tCLUSTER\tDEPLOY_ID\t"}
	// TODO: Can we find the failed object?
	for _, item := range deploys {
		ads := item.ADS
		pattern := "%s\t%s\t%s\t%s\t%s\t"
		status := "\x1b[32mDeployed\x1b[0m"
		if !item.Success {
			status = "\x1b[31mFailed\x1b[0m"
		}
		result := fmt.Sprintf(pattern, status, ads.Name, ads.Namespace, ads.Cluster, item.DeployId)
		results = append(results, result)
	}

	if len(deploys) > 0 {
		printDeployResults(results)
	}

	return nil
}

func printDeployResults(results []string) {
	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)
	for _, result := range results {
		fmt.Fprintln(w, result)
	}
	w.Flush()
}

func createApplicationIds(apps []string) []ApplicationId {
	applicationIds := []ApplicationId{}
	for _, app := range apps {
		envApp := strings.Split(app, "/")
		applicationIds = append(applicationIds, ApplicationId{
			Environment: envApp[0],
			Application: envApp[1],
		})
	}
	return applicationIds
}
