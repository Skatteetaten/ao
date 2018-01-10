package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

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
	user, _ := os.LookupEnv("USER")
	loginCmd.Flags().StringVarP(&flagUserName, "username", "u", user, "the username to log in with, standard is $USER")
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
