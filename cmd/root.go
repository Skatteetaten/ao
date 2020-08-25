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

const (
	bashCompletionFunc = `__ao_parse()
{
    local ao_output out
    if ao_output=$(ao $@ --no-headers 2>/dev/null); then
        out=($(echo "${ao_output}" | awk '{print $1}'))
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
    fi
}

__custom_func() {
    case ${last_command} in
        ao_edit | ao_get_file | ao_delete_file | ao_set | ao_unset)
            __ao_parse get files
            return
            ;;
        ao_deploy | ao_get_spec)
            __ao_parse get all --list
            return
            ;;
        ao_vault_edit | ao_vault_delete-secret | ao_vault_rename-secret)
            __ao_parse vault get --list
            return
            ;;
        ao_vault_delete | ao_vault_rename | ao_vault_permissions)
            __ao_parse vault get --only-vaults
            return
            ;;
        *)
            ;;
    esac
}
`
)

const rootLong = `A command line interface for the Boober API.
  * Deploy one or more ApplicationDeploymentRef (environment/application) to one or more clusters
  * Manage AuroraConfig remotely
  * Support modifying AuroraConfig locally
  * Manage vaults and secrets`

var (
	pFlagLogLevel  string
	pFlagPrettyLog bool
	pFlagToken     string
	pFlagRefName   string
	pFlagNoHeader  bool

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
	Use:   "ao",
	Short: "Aurora OpenShift CLI",
	Long:  rootLong,
	// Cannot use custom bash completion until https://github.com/spf13/cobra/pull/520 has been merged
	// BashCompletionFunction: bashCompletionFunc,
	PersistentPreRunE: initialize,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&pFlagLogLevel, "log", "l", "fatal", "Set log level. Valid log levels are [info, debug, warning, error, fatal]")
	RootCmd.PersistentFlags().BoolVarP(&pFlagPrettyLog, "pretty", "p", false, "Pretty print json output for log")
	RootCmd.PersistentFlags().StringVarP(&pFlagToken, "token", "t", "", "OpenShift authorization token to use for remote commands, overrides login")
	RootCmd.PersistentFlags().StringVarP(&pFlagRefName, "ref", "", "", "Set git ref name, does not affect vaults")
	RootCmd.PersistentFlags().BoolVarP(&pFlagNoHeader, "no-headers", "", false, "Print tables without headers")
	RootCmd.PersistentFlags().MarkHidden("no-headers")
}

func initialize(cmd *cobra.Command, args []string) error {

	// Setting output for cmd.Print methods
	cmd.SetOutput(os.Stdout)
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
		logrus.Debugf("ao config is saved with another versjon. AO-version: %s, saved version: %s", config.Version, aoConfig.FileAOVersion)
		fmt.Printf("\nThe current ao config is made with an older version of ao.\n")
		message := "Do you want to recreate the ao config with default values (recommended)?"
		update := prompt.Confirm(message, true)
		if update {
			RecreateConfig(cmd, args)
			if aoConfig, err = config.LoadConfigFile(ConfigLocation); err != nil {
				logrus.Error(fmt.Errorf("Could not load config after recreate: %w", err))
			}
		} else {
			fmt.Printf("\nNB: Using the older config may cause errors. \nIf you experience this, try running command \"ao adm recreate-config\".\n\n")
		}
	}

	if flagAuroraConfig == "" && flagCheckoutAffiliation == "" {
		commandsWithoutAffiliation := []string{"version", "login", "logout", "adm", "update"}
		if containsNone(cmd.CommandPath(), commandsWithoutAffiliation) && aoConfig.Affiliation == "" {
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

	api := &client.APIClient{
		Affiliation: aoConfig.Affiliation,
		Host:        apiCluster.BooberURL,
		GoboHost:    apiCluster.GoboURL,
		Token:       apiCluster.Token,
		RefName:     aoConfig.RefName,
	}

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
