package newappcmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
)

const UsageString = "Usage: aoc new-app <appname>"
const AppnameOrArtictNeeded = "Have to specify either an app-name or an artifact name"
const MissingArgumentFormat = "Missing %v"
const InteractiveNoFlags = "No flags or arguments allowed for interactive run"
const DeployNeedVersion = "Need to have a version for deployment type deploy"
const appExistsError = "Error: App exists"

type NewappcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (newappcmdClass *NewappcmdClass) getAffiliation() (affiliation string) {
	if newappcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = newappcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (newappcmdClass *NewappcmdClass) NewappCommand(args []string, artifactid string, cluster string, env string, groupid string, interactive bool, outputFolder string, deployentType string, version string) (output string, err error) {
	err = validateNewappCommand(args, artifactid, cluster, env, groupid, interactive, outputFolder, deployentType, version)
	if err != nil {
		return "", err
	}
	return

}

func validateNewappCommand(args []string, artifactid string, cluster string, env string, groupid string, interactive bool, outputFolder string, deploymentType string, version string) (err error) {
	// Check for interactive, then no other parameters should be given
	var appname = ""
	if len(args) > 2 {
		err = errors.New(UsageString)
		return err
	}
	if len(args) > 1 {
		appname = args[1]
	}

	if interactive {
		if appname != "" || artifactid != "" || cluster != "" || env != "" || groupid != "" || outputFolder != "" || deploymentType != "" || version != "" {
			err = errors.New(InteractiveNoFlags)
			return err
		}
	} else {
		// Check that we have either an app-name or an artifact id
		if appname == "" && artifactid == "" {
			err = errors.New(AppnameOrArtictNeeded)
			return err
		}

		// Check that we have a version if type is deployment
		if deploymentType == "deploy" {
			if version == "" {
				err = errors.New(DeployNeedVersion)
				return err
			}
		}

		// Check that we have cluster
		if cluster == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "cluster"))
			return err
		}

		if env == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "env"))
			return err
		}

		if groupid == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "groupid"))
			return err
		}

		if deploymentType == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "deploymentType"))
			return err
		}
	}
	return
}
