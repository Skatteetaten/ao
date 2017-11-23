package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"os"
)

var (
	flagUserName string
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
}

func Login(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Please specify AuroraConfig to log in to")
	}

	if args[0] != "" {
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

	return config.WriteConfig(*AO, ConfigLocation)
}
