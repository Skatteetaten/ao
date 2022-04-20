package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	architect "github.com/skatteetaten/architect/v2/pkg/build"
	architectConfig "github.com/skatteetaten/architect/v2/pkg/config"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const OpenshiftRegistryFormat = "default-route-openshift-image-registry.apps.%s.paas.skead.no"

var (
	devCmd = &cobra.Command{
		Use:   "dev",
		Short: "Perform dev operations",
		// TODO: this should possibly have its own sections in help, development commands?
		Annotations: map[string]string{"type": "actions"},
	}

	rolloutCmd = &cobra.Command{
		Use:   "rollout <applicationdeploymentRef> <Leveransepakke file>",
		Short: "Rollout leveransepakke to openshift for development type",
		// TODO: this should possibly have its own sections in help, development commands?
		Annotations: map[string]string{"type": "actions"},
		RunE:        Rollout,
	}
)

func init() {
	devCmd.AddCommand(rolloutCmd)
	RootCmd.AddCommand(devCmd)
}

// TODO: this does not build for windows, because of architect. What to do?
func Rollout(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	leveransePakkeFile := args[1]
	if err := verifyFileExists(leveransePakkeFile); err != nil {
		return err
	}

	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]
	matches := auroraconfig.FindMatches(search, fileNames.GetApplicationDeploymentRefs(), false)
	if len(matches) == 0 {
		return errors.Errorf("No matches for %s", search)
	} else if len(matches) > 1 {
		return errors.Errorf("Search matched more than one file. Search must be more specific.\n%v", matches)
	}

	allSpecs, err := DefaultAPIClient.GetAuroraDeploySpec(matches, !flagNoDefaults, flagIgnoreErrors)
	if err != nil {
		return err
	}
	spec := allSpecs[0]
	deployType := spec.Get("type")
	if deployType != "development" {
		return errors.New("Needs to be development")
	}

	name := spec.Name()
	cluster := spec.Cluster()
	appType := spec.GetString("applicationPlatform")
	adID := spec.GetString("applicationDeploymentId")
	baseImageName := spec.GetString("/baseImage/name")
	baseImageVersion := spec.GetString("/baseImage/version")
	namespace, err := DefaultAPIClient.GetNamespace(adID)
	if err != nil {
		return err
	}

	conf := getArchitectBuildConfiguration(appType, baseImageName, baseImageVersion, cluster, namespace, name, leveransePakkeFile)
	architect.BuildBinary(conf)

	return nil
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

func getArchitectBuildConfiguration(applicationPlatform string, baseImageName string, baseImageVersion string, cluster string, namespace string, name string, file string) architect.Configuration {
	appType := parseApplicationType(applicationPlatform)
	outputRepository := fmt.Sprintf("%s/%s", namespace, name)
	baseImageFullName := fmt.Sprintf("aurora/%s", baseImageName)

	registry := getClusterRegistryUrl(cluster)

	token := AOSession.Tokens[cluster]

	conf := architect.Configuration{
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

	return conf
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
