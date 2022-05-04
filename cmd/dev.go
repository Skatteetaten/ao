// Architect will not build for windows, therefore we exclude the whole command for windows
//go:build !windows

package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/deploymentspec"
	"github.com/skatteetaten/ao/pkg/prompt"
	architect "github.com/skatteetaten/architect/v2/pkg/build"
	architectConfig "github.com/skatteetaten/architect/v2/pkg/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const OpenshiftRegistryFormat = "default-route-openshift-image-registry.apps.%s.paas.skead.no"

var flagDevNoPrompt bool

const ExampleRollout = ` Given the following AuroraConfig:
    - about.json
    - foobar.json
    - bar.json
    - foo/about.json
    - foo/bar.json
		- foo/foobar.json
		- ref/about.json
    - ref/bar.json

  Given leveransepakke file: whoami-Leveransepakke.zip 

  # Fuzzy matching: rollout foo/bar and foo/foobar
  ao dev rollout fo/ba whoami-Leveransepakke.zip

  # Exact matching: rollout foo/bar
  ao dev rollout foo/bar whoami-Leveransepakke.zip
`

var (
	devCmd = &cobra.Command{
		Use:         "dev",
		Short:       "Perform dev operations",
		Annotations: map[string]string{"type": "development"},
	}

	rolloutCmd = &cobra.Command{
		Use:     "rollout <applicationdeploymentRef> <Leveransepakke file>",
		Short:   "Rollout leveransepakke to openshift for development type",
		Example: ExampleRollout,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Rollout(cmd, args, architect.BuildBinary, DefaultAPIClient)
		},
	}
)

func init() {
	devCmd.AddCommand(rolloutCmd)
	devCmd.Flags().BoolVarP(&flagDevNoPrompt, "yes", "y", false, "Suppress prompts and accept deployment(s)")
	devCmd.Flags().BoolVarP(&flagDevNoPrompt, "no-prompt", "", false, "Suppress prompts and accept deployment(s)")
	RootCmd.AddCommand(devCmd)
}

type DevRolloutClient interface {
	GetFileNames() (auroraconfig.FileNames, error)
	GetAuroraDeploySpec(applications []string, defaults bool, ignoreErrors bool) ([]deploymentspec.DeploymentSpec, error)
}

func Rollout(cmd *cobra.Command, args []string, buildBinaryFunc func(c architect.Configuration), rolloutClient DevRolloutClient) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	leveransePakkeFile := args[1]
	if err := verifyFileExists(leveransePakkeFile); err != nil {
		return err
	}

	fileNames, err := rolloutClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]

	matchedAdr, err := getFuzzyMatchedAdr(search, fileNames)
	if err != nil {
		return err
	}

	spec, err := getSpec(*matchedAdr, rolloutClient)
	if err != nil {
		return err
	}

	deployType := spec.Get("type")
	if deployType != "development" {
		return fmt.Errorf("You need to specify type=development in AuroraConfig. Current type=%s", deployType)
	}

	if !flagDevNoPrompt {
		shouldContinue := prompt.Confirm(fmt.Sprintf("Do you wish to rollout ApplicationDeployment %s", *matchedAdr), true)
		if !shouldContinue {
			return errors.New("Aborted...")
		}
	}

	conf := getArchitectBuildConfiguration(spec, leveransePakkeFile)

	cmd.Println("Building image.. If you wish to see the details you can run the command with `-l info` flag.")
	buildBinaryFunc(conf)

	return nil
}

func getSpec(matchedAdr string, rolloutClient DevRolloutClient) (deploymentspec.DeploymentSpec, error) {
	allSpecs, err := rolloutClient.GetAuroraDeploySpec([]string{matchedAdr}, true, false)
	if err != nil {
		return nil, err
	}

	return allSpecs[0], nil
}

func getFuzzyMatchedAdr(search string, fileNames auroraconfig.FileNames) (*string, error) {
	adrs := auroraconfig.FindMatches(search, fileNames.GetApplicationDeploymentRefs(), false)
	if len(adrs) == 0 {
		return nil, errors.Errorf("No matches for %s", search)
	} else if len(adrs) > 1 {
		return nil, errors.Errorf("Search matched more than one file. Search must be more specific.\nMatched: %v", adrs)
	}

	return &adrs[0], nil
}

func verifyFileExists(file string) error {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("the provided path and name of the leveransepakke file does not exist")
	}

	return nil
}

func getClusterRegistryUrl(cluster string) string {
	return fmt.Sprintf(OpenshiftRegistryFormat, cluster)
}

func getArchitectBuildConfiguration(spec deploymentspec.DeploymentSpec, file string) architect.Configuration {
	appType := parseApplicationType(spec.GetString("applicationPlatform"))
	name := spec.Name()
	cluster := spec.Cluster()
	baseImageName := spec.GetString("/baseImage/name")
	baseImageVersion := spec.GetString("/baseImage/version")
	namespace := spec.GetString("namespace")

	outputRepository := fmt.Sprintf("%s/%s", namespace, name)
	baseImageFullName := fmt.Sprintf("aurora/%s", baseImageName)

	registry := getClusterRegistryUrl(cluster)

	token := AOSession.Tokens[cluster]

	return architect.Configuration{
		File:             file,
		BaseImageName:    baseImageFullName,
		ApplicationType:  appType,
		OutputRepository: outputRepository,
		TagWith:          "latest",
		BaseImageVersion: baseImageVersion,
		PushRegistry:     registry,
		PullRegistry:     "https://container-registry-internal-private-pull.aurora.skead.no",
		Version:          "latest",
		PushToken:        token,
		PushUsername:     "user",
	}
}

func parseApplicationType(applicationType string) (appType architectConfig.ApplicationType) {
	switch strings.ToLower(applicationType) {
	case "java":
		appType = architectConfig.JavaLeveransepakke
	case "web":
		appType = architectConfig.NodeJsLeveransepakke
	case "doozer":
		appType = architectConfig.DoozerLeveranse
	default:
		appType = architectConfig.JavaLeveransepakke
	}

	return
}
