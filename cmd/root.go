package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/prompt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/log"
	"github.com/spf13/cobra"
)

const rootLong = `A command line interface for the Boober API.
  * Deploy one or more ApplicationDeploymentRef (environment/application) to one or more clusters
  * Manage AuroraConfig remotely
  * Support modifying AuroraConfig locally
  * Manage vaults and secrets`

var (
	pFlagLogLevel             string
	pFlagPrettyLog            bool
	pFlagToken                string
	pFlagRefName              string
	pFlagNoHeader             bool
	pFlagAnswerRecreateConfig string

	// DefaultAPIClient will use APICluster from ao config as default values
	// if persistent token and/or server api url is specified these will override default values
	DefaultAPIClient *client.APIClient
	// AO holds the config og ao
	AO *config.AOConfig
	// ConfigLocation is the location of the config
	ConfigLocation string
)

// RootCmd is the root of the entire `ao` cli command structure
var RootCmd = &cobra.Command{
	Use:               "ao command",
	Short:             "Aurora OpenShift CLI",
	Long:              rootLong,
	PersistentPreRunE: initialize,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&pFlagLogLevel, "log", "l", "fatal", "Set log level. Valid log levels are [info, debug, warning, error, fatal]")
	RootCmd.PersistentFlags().BoolVarP(&pFlagPrettyLog, "pretty", "p", false, "Pretty print json output for log")
	RootCmd.PersistentFlags().StringVarP(&pFlagToken, "token", "t", "", "OpenShift authorization token to use for remote commands, overrides login")
	RootCmd.PersistentFlags().StringVar(&pFlagRefName, "ref", "", "Set git ref name, does not affect vaults")
	RootCmd.PersistentFlags().BoolVar(&pFlagNoHeader, "no-headers", false, "Print tables without headers")
	RootCmd.PersistentFlags().MarkHidden("no-headers")
	RootCmd.PersistentFlags().StringVar(&pFlagAnswerRecreateConfig, "autoanswer-recreate-config", "", "Set automatic response for ao config question [y, n]")
}

func initialize(cmd *cobra.Command, args []string) error {

	// Setting output for cmd.Print methods
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	// Errors will be printed from main
	cmd.SilenceErrors = true
	// Disable print usage when an error occurs
	cmd.SilenceUsage = true

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	ConfigLocation = filepath.Join(home, ".ao.json")

	err = setLogging(pFlagLogLevel, pFlagPrettyLog)
	if err != nil {
		return err
	}

	aoConfig, err := config.LoadConfigFile(ConfigLocation)
	if err != nil {
		logrus.Error(err)
	}

	if aoConfig == nil {
		logrus.Info("Creating new config")
		aoConfig = &config.DefaultAOConfig
		aoConfig.InitClusters()
		aoConfig.SelectAPICluster()
		err = config.WriteConfig(*aoConfig, ConfigLocation)
		if err != nil {
			return err
		}
	} else if aoConfig.FileAOVersion != config.Version {
		logrus.Debugf("ao config file is saved with another versjon. AO-version: %s, saved version: %s", config.Version, aoConfig.FileAOVersion)

		if update() {
			RecreateConfig(cmd, args)
			if aoConfig, err = config.LoadConfigFile(ConfigLocation); err != nil {
				logrus.Error(fmt.Errorf("Could not load config after recreate: %w", err))
			}
			if pFlagAnswerRecreateConfig == "" {
				fmt.Printf("\nThe ao configuration settings file was updated to match the current ao version.\n\n")
			}
		} else {
			if pFlagAnswerRecreateConfig == "" {
				fmt.Printf("\nNB: Using the ao configuration settings file created for another ao version may cause errors. \nIf you experience errors, try running command \"ao adm recreate-config\".\n\n")
			}
		}
	}

	if flagAuroraConfig == "" && flagCheckoutAffiliation == "" {
		commandsWithoutAffiliation := []string{"version", "login", "logout", "adm", "update"}
		if !strings.Contains(cmd.Name(), cobra.ShellCompRequestCmd) && containsNone(cmd.CommandPath(), commandsWithoutAffiliation) && aoConfig.Affiliation == "" {
			return errors.New("no affiliations is set, please login")
		}
	}

	apiCluster := aoConfig.Clusters[aoConfig.APICluster]
	if apiCluster == nil {
		if !strings.Contains(cmd.CommandPath(), "adm") {
			return errors.Errorf("api cluster %s is not available. Check config", aoConfig.APICluster)
		}
		apiCluster = &config.Cluster{}
	}

	api := client.NewAPIClient(apiCluster.BooberURL, apiCluster.GoboURL, apiCluster.Token, aoConfig.Affiliation, aoConfig.RefName, client.CreateUUID().String())

	if aoConfig.Localhost {
		// TODO: Move to config?
		api.Host = "http://localhost:8080"
		api.GoboHost = "http://localhost:8080"
	}

	if pFlagRefName != "" {
		api.RefName = pFlagRefName
	}

	if pFlagToken != "" {
		api.Token = pFlagToken
	}

	AO, DefaultAPIClient = aoConfig, api

	return nil
}

func update() bool {
	// ask for update if none of the flags "token" or "autoanswer-recreate-config" are set
	ask := pFlagToken == "" && pFlagAnswerRecreateConfig == ""

	if ask {
		fmt.Printf("\nIt looks like ao have been updated to another version.\n")
		message := "Do you want to recreate the ao configuration settings file with default values (recommended)?"
		return prompt.Confirm(message, true)
	}

	return strings.ToLower(pFlagAnswerRecreateConfig) != "n"
}

func containsNone(value string, list []string) bool {
	none := true
	for _, v := range list {
		if strings.Contains(value, v) {
			none = false
		}
	}
	return none
}

func setLogging(level string, pretty bool) error {
	logrus.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	if pretty {
		logrus.SetFormatter(&log.PrettyFormatter{})
	}

	return nil
}
