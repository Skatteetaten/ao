package cmd

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

// Change this for new releases of the Boober API
const supportedApiVersion = 2

var (
	flagUserName   string
	flagLocalhost  bool
	flagApiCluster string
)

var loginCmd = &cobra.Command{
	Use:   "login <AuroraConfig>",
	Short: "Login to all available OpenShift clusters",
	RunE:  Login,
}

func init() {
	RootCmd.AddCommand(loginCmd)
	var username string
	if runtime.GOOS == "windows" {
		user, err := user.Current()
		if err != nil {
			logrus.Fatal("Unable to get current User info: " + err.Error())
		}
		if strings.Contains(user.Username, "\\") {
			parts := strings.Split(user.Username, "\\")
			if len(parts) > 0 {
				username = parts[1]
			}
		}
	} else {
		username, _ = os.LookupEnv("USER")
	}

	loginCmd.Flags().StringVarP(&flagUserName, "username", "u", username, "the username to log in with, standard is current user")
	loginCmd.Flags().BoolVarP(&flagLocalhost, "localhost", "", false, "set api to localhost")
	loginCmd.Flags().MarkHidden("localhost")
	loginCmd.Flags().StringVarP(&flagApiCluster, "apicluster", "", "", "select specified API cluster")
	loginCmd.Flags().MarkHidden("apicluster")
}

func Login(cmd *cobra.Command, args []string) error {
	if len(args) != 1 && AO.Affiliation == "" { // Dont demand an AuroraConfig if we have one in the config
		return errors.New("Please specify AuroraConfig to log in to")
	}
	if len(args) == 1 {
		AO.Affiliation = args[0]
	}

	var password string
	for _, c := range AO.Clusters {
		if !c.Reachable || c.HasValidToken() {
			continue
		}
		if password == "" {
			password = prompt.Password()
		}
		token, err := config.GetToken(c.Url, flagUserName, password)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"url":      c.Url,
				"userName": flagUserName,
			}).Fatal(err)
		}
		c.Token = token
	}

	AO.Update(false)

	var supressAffiliationCheck bool

	if flagApiCluster != "" {
		AO.APICluster = flagApiCluster
		// Can't check for legal affiliations in new cluster, so dont bother
		supressAffiliationCheck = true
	}

	acn, err := DefaultApiClient.GetAuroraConfigNames()
	if err != nil {
		if !AO.Localhost {
			return err
		}
		supressAffiliationCheck = true
	}

	if !supressAffiliationCheck {
		var apiVersion int
		clientConfig, err := DefaultApiClient.GetClientConfig()
		if err != nil {
			return err
		}
		apiVersion, err = strconv.Atoi(clientConfig.ApiVersion)
		if err != nil {
			return errors.Wrap(err, "Expected apiVersion to be an int")
		}

		if apiVersion != supportedApiVersion {
			var grade string
			if apiVersion < supportedApiVersion {
				grade = "downgrade"
			} else {
				grade = "upgrade"
			}
			message := fmt.Sprintf("This version of AO does not support Boober with api version %v, you need to %v.", apiVersion, grade)

			return errors.New(message)
		}

		var found bool
		for _, affiliation := range *acn {
			if affiliation == AO.Affiliation {
				found = true
				break
			}
		}
		if !found {
			err := errors.New("Illegal affiliation: " + AO.Affiliation)
			return err
		}
	}

	AO.Localhost = flagLocalhost
	return config.WriteConfig(*AO, ConfigLocation)
}
