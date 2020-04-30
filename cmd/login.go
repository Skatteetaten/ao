package cmd

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"

	"ao/pkg/config"
	"ao/pkg/prompt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Change this for new releases of the Boober API
const supportedAPIVersion = 2

var (
	flagPassword   string
	flagUserName   string
	flagLocalhost  bool
	flagAPICluster string
)

var loginCmd = &cobra.Command{
	Use:     "login <AuroraConfig>",
	Short:   "Login to all available OpenShift clusters",
	PreRunE: PreLogin,
	RunE:    Login,
	PostRun: PostLogin,
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
	loginCmd.Flags().StringVarP(&flagPassword, "password", "", "", "the password to log in with, if not set will prompt.  Should only be used in combination with a capturing function to avoid beeing shown in history files")
	loginCmd.Flags().BoolVarP(&flagLocalhost, "localhost", "", false, "set api to localhost")
	loginCmd.Flags().MarkHidden("localhost")
	loginCmd.Flags().StringVarP(&flagAPICluster, "apicluster", "", "", "select specified API cluster")
	loginCmd.Flags().MarkHidden("apicluster")
}

// PreLogin performs pre command validation checks for the `login` cli command
func PreLogin(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Please specify AuroraConfig to log in to")
	}
	if len(args) == 1 {
		AO.Affiliation = args[0]
	}

	var password string
	if flagPassword != "" {
		password = flagPassword
	}
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

	return nil
}

// Login performs main part of the `login` cli command
func Login(cmd *cobra.Command, args []string) error {
	if AO.Localhost != flagLocalhost {
		AO.Localhost = flagLocalhost
	}

	if flagAPICluster != "" {
		if _, ok := AO.Clusters[flagAPICluster]; !ok {
			return errors.Errorf("%s is not a valid cluster option. Choose between %v", flagAPICluster, AO.AvailableClusters)
		}
		AO.APICluster = flagAPICluster
	}

	cluster := AO.Clusters[AO.APICluster]
	DefaultAPIClient.Token = cluster.Token

	host := cluster.BooberUrl
	gobohost := cluster.GoboUrl

	if AO.Localhost {
		host = "http://localhost:8080"
	}
	DefaultAPIClient.Host = host
	DefaultAPIClient.GoboHost = gobohost

	acn, err := DefaultAPIClient.GetAuroraConfigNames()
	if err != nil {
		return err
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

	var apiVersion int
	clientConfig, err := DefaultAPIClient.GetClientConfig()
	if err != nil {
		return err
	}

	apiVersion = clientConfig.APIVersion
	if apiVersion != supportedAPIVersion {
		var grade string
		if apiVersion < supportedAPIVersion {
			grade = "downgrade"
		} else {
			grade = "upgrade"
		}
		message := fmt.Sprintf("This version of AO does not support Boober with api version %v, you need to %v.", apiVersion, grade)
		return errors.New(message)
	}

	AO.Update(false)
	return config.WriteConfig(*AO, ConfigLocation)
}

// PostLogin shows results at the end of performing the `login` cli command
func PostLogin(cmd *cobra.Command, args []string) {

	PrintClusters(cmd, true)
	if AO.RefName != "" && AO.RefName != "master" {
		fmt.Printf("\nrefName=%s in AO configurations file. Consider running command \"ao adm update-ref <refName>\" if this is incorrect\n", AO.RefName)
	}
	fmt.Printf("\nConsider running command \"ao adm update-clusters\" if cluster information above looks incorrect \n")
}
