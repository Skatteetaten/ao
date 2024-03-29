package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/skatteetaten/ao/pkg/service"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

var applicationDeploymentRedeployCmd = &cobra.Command{
	Use:   "redeploy <applicationDeploymentRef>",
	Short: "Redeploy running application deployment(s) with the given reference",
	RunE:  redeployApplicationDeployment,
}

func init() {
	applicationDeploymentCmd.AddCommand(applicationDeploymentRedeployCmd)
	applicationDeploymentRedeployCmd.Flags().StringVarP(&flagCluster, "cluster", "c", "", "Limit redeploy to given cluster name")
	applicationDeploymentRedeployCmd.Flags().BoolVarP(&flagNoPrompt, "yes", "y", false, "Suppress prompts and accept redeploy")
	applicationDeploymentRedeployCmd.Flags().BoolVarP(&flagNoPrompt, "no-prompt", "", false, "Suppress prompts and accept redeploy")
	applicationDeploymentRedeployCmd.Flags().StringArrayVarP(&flagExcludes, "exclude", "e", []string{}, "Select applications or environments to exclude from redeploy")

	applicationDeploymentRedeployCmd.Flags().BoolVarP(&flagNoPrompt, "force", "f", false, "Suppress prompts")
	applicationDeploymentRedeployCmd.Flags().MarkHidden("force")
	applicationDeploymentRedeployCmd.Flags().StringVarP(&flagAuroraConfig, "affiliation", "", "", "Overrides the logged in affiliation")
	applicationDeploymentRedeployCmd.Flags().MarkHidden("affiliation")
}

func redeployApplicationDeployment(cmd *cobra.Command, args []string) error {
	apiCluster := flagCluster
	if len(strings.TrimSpace(pFlagAPICluster)) > 0 {
		apiCluster = strings.TrimSpace(pFlagAPICluster)
	}

	if len(args) > 2 || len(args) < 1 {
		return cmd.Usage()
	}

	err := validateRedeployParams()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	auroraConfigName := AOSession.AuroraConfig
	if flagAuroraConfig != "" {
		auroraConfigName = flagAuroraConfig
	}

	apiClient, err := getAPIClient(auroraConfigName, pFlagToken, apiCluster)
	if err != nil {
		return err
	}

	applications, err := service.GetApplications(apiClient, search, flagExcludes)
	if err != nil {
		return err
	} else if len(applications) == 0 {
		return errors.New("No applications to redeploy")
	}

	filteredDeploymentSpecs, err := service.GetFilteredDeploymentSpecs(apiClient, applications, flagCluster)
	if err != nil {
		return err
	}

	err = checkForDuplicateSpecs(filteredDeploymentSpecs)
	if err != nil {
		return err
	}

	activeDeploymentSpecs, err := getDeployedDeploymentSpecs(getApplicationDeploymentClient, filteredDeploymentSpecs, auroraConfigName, pFlagToken)
	if err != nil {
		return err
	} else if len(activeDeploymentSpecs) == 0 {
		return errors.New("No applications to redeploy")
	}

	partitions, err := createDeploySpecPartitions(auroraConfigName, pFlagToken, AOConfig.Clusters, activeDeploymentSpecs)
	if err != nil {
		return err
	}

	if !getRedeployConfirmation(flagNoPrompt, activeDeploymentSpecs, cmd.OutOrStdout()) {
		return errors.New("No applications to redeploy")
	}

	result, unsuccessfulErr := deployToReachableClusters(getApplicationDeploymentClient, partitions, make(map[string]string))

	printDeployResult(result, cmd.OutOrStdout())

	return unsuccessfulErr
}

func checkForDuplicateSpecs(deploymentSpecs []deploymentspec.DeploymentSpec) error {
	if len(deploymentSpecs) > 1 {
		for i := 0; i < len(deploymentSpecs)-1; i++ {
			for j := i + 1; j < len(deploymentSpecs); j++ {
				if deploymentSpecs[i].Name() == deploymentSpecs[j].Name() &&
					deploymentSpecs[i].Environment() == deploymentSpecs[j].Environment() &&
					deploymentSpecs[i].Cluster() == deploymentSpecs[j].Cluster() {
					return fmt.Errorf("Can not redeploy, since there are several aurora config specs for (cluster env app): %v %v %v",
						deploymentSpecs[i].Cluster(),
						deploymentSpecs[i].Environment(),
						deploymentSpecs[i].Name())
				}
			}
		}
	}
	return nil
}

func validateRedeployParams() error {
	if flagCluster != "" {
		if _, exists := AOConfig.Clusters[flagCluster]; !exists {
			return errors.New(fmt.Sprintf("No such cluster %s", flagCluster))
		}
	}

	return nil
}

func getRedeployConfirmation(force bool, filteredDeploymentSpecs []deploymentspec.DeploymentSpec, out io.Writer) bool {
	header, rows := GetDeploySpecTable(filteredDeploymentSpecs, "")
	DefaultTablePrinter(header, rows, out)

	shouldDeploy := true
	if !force {
		defaultAnswer := len(rows) == 1
		message := fmt.Sprintf("Do you want to redeploy %d application(s) in affiliation %s?", len(rows), AOSession.AuroraConfig)
		shouldDeploy = prompt.Confirm(message, defaultAnswer)
	}

	return shouldDeploy
}
