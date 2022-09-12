package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/skatteetaten/ao/pkg/session"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/log"
	"github.com/spf13/cobra"
)

const rootLong = `A command line interface for the Aurora API.
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
	pFlagAPICluster           string
	pFlagAnswerRecreateConfig string // deprecated

	// DefaultAPIClient will use APICluster from ao config as default values
	// if persistent token and/or server api url is specified these will override default values
	DefaultAPIClient *client.APIClient
	// AOConfig holds the ao config
	AOConfig *config.AOConfig
	// CustomConfigLocation is the location of an optional config file
	CustomConfigLocation string
	// AO holds the ao session
	AOSession *session.AOSession
	// SessionFileLocation is the location of the file holding session data for the login session
	SessionFileLocation string
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
	RootCmd.PersistentFlags().StringVarP(&pFlagAPICluster, "apicluster", "", "", "specify API cluster for this command, persistent when used with login")
	RootCmd.PersistentFlags().MarkHidden("no-headers")
	RootCmd.PersistentFlags().StringVar(&pFlagAnswerRecreateConfig, "autoanswer-recreate-config", "", "deprecated")
	RootCmd.PersistentFlags().MarkHidden("autoanswer-recreate-config")
}

func initialize(cmd *cobra.Command, args []string) error {

	// Setting output for cmd.Print methods
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	// Errors will be printed from main
	cmd.SilenceErrors = true
	// Disable print usage when an error occurs
	cmd.SilenceUsage = true

	if strings.Contains(cmd.Name(), cobra.ShellCompRequestCmd) {
		// cmd is an auto completion call
		return nil
	}

	err := setLogging(pFlagLogLevel, pFlagPrettyLog)
	if err != nil {
		return err
	}
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("Error while resolving home dir: %w", err)
	}
	CustomConfigLocation = filepath.Join(home, ".ao-config.json")
	SessionFileLocation = filepath.Join(home, ".ao-session.json")

	aoConfig := config.LoadOrCreateAOConfig(CustomConfigLocation)

	aoSession, err := session.LoadOrCreateAOSessionFile(SessionFileLocation, aoConfig)
	if err != nil {
		return err
	}

	if flagAuroraConfig == "" && flagCheckoutAuroraconfig == "" {
		commandsWithoutAffiliation := []string{"version", "login", "logout", "adm", "update"}
		if containsNone(cmd.CommandPath(), commandsWithoutAffiliation) && aoSession.AuroraConfig == "" {
			return errors.New("No affiliations is set. Please log in.")
		}
	}

	var apiClusterName string
	if len(strings.TrimSpace(pFlagAPICluster)) > 0 {
		apiClusterName = strings.TrimSpace(pFlagAPICluster)
	} else {
		apiClusterName = aoSession.APICluster
	}
	apiCluster := aoConfig.Clusters[apiClusterName]

	if apiCluster == nil {
		if !strings.Contains(cmd.CommandPath(), "adm") {
			return errors.Errorf("api cluster %s is not available. Try again later.", apiClusterName)
		}
		apiCluster = &config.Cluster{}
	}
	apiToken := aoSession.Tokens[apiClusterName]

	api := client.NewAPIClient(apiCluster.BooberURL, apiCluster.GoboURL, apiToken, aoSession.AuroraConfig, aoSession.RefName, client.CreateUUID().String())

	if aoSession.Localhost {
		api.Host = "http://localhost:8080"
		api.GoboHost = "http://localhost:8080"
	}

	if pFlagRefName != "" {
		api.RefName = pFlagRefName
	}

	if pFlagToken != "" {
		api.Token = pFlagToken
	}

	AOConfig, AOSession, DefaultAPIClient = aoConfig, aoSession, api

	return nil
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
