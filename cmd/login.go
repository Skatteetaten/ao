package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/session"
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
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
}

// PreLogin performs pre command validation checks for the `login` cli command
func PreLogin(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Please specify AuroraConfig to log in to")
	}
	if len(args) == 1 {
		AOSession.AuroraConfig = args[0]
	}

	var password string
	if flagPassword != "" {
		password = flagPassword
	}
	for _, c := range AOConfig.Clusters {
		if !c.Reachable || c.IsValidToken(AOSession.Tokens[c.Name]) {
			continue
		}
		if password == "" {
			password = prompt.Password()
		}
		token, err := config.GetToken(c.LoginURL, flagUserName, password)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"url":      c.URL,
				"userName": flagUserName,
			}).Fatal(err)
		}
		AOSession.Tokens[c.Name] = token
	}

	return nil
}

// Login performs main part of the `login` cli command
func Login(cmd *cobra.Command, args []string) error {
	if AOSession.Localhost != flagLocalhost {
		AOSession.Localhost = flagLocalhost
	}

	if flagAPICluster != "" {
		if _, ok := AOConfig.Clusters[flagAPICluster]; !ok {
			return errors.Errorf("%s is not a valid cluster option. Choose between %v", flagAPICluster, AOConfig.AvailableClusters)
		}
		AOSession.APICluster = flagAPICluster
	}

	cluster := AOConfig.Clusters[AOSession.APICluster]
	DefaultAPIClient.Token = AOSession.Tokens[cluster.Name]

	host := cluster.BooberURL
	gobohost := cluster.GoboURL

	if AOSession.Localhost {
		host = "http://localhost:8080"
	}
	DefaultAPIClient.Host = host
	DefaultAPIClient.GoboHost = gobohost

	acn, err := DefaultAPIClient.GetAuroraConfigNames()
	if err != nil {
		return fmt.Errorf("While loading aurora config names: %w", err)
	}
	var found bool
	for _, auroraConfigName := range *acn {
		if auroraConfigName == AOSession.AuroraConfig {
			found = true
			break
		}
	}
	if !found {
		err := errors.New("Illegal aurora config: " + AOSession.AuroraConfig)
		return err
	}

	var apiVersion int
	clientConfig, err := DefaultAPIClient.GetClientConfig()
	if err != nil {
		return fmt.Errorf("While getting client config: %w", err)
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

	if err = session.WriteAOSession(*AOSession, SessionFileLocation); err != nil {
		return err
	}

	aoUpdated, err := AOConfig.Update(false)
	cmd.Annotations = make(map[string]string)
	if aoUpdated {
		cmd.Annotations["Updated"] = "true"
	}
	if err != nil {
		logrus.Debug(err)
	}

	return nil
}

// PostLogin shows results at the end of performing the `login` cli command
func PostLogin(cmd *cobra.Command, args []string) {
	if cmd.Annotations["Updated"] == "true" {
		fmt.Println("AO was updated.")
	} else {
		PrintClusters(cmd, true)
		if AOSession.RefName != "" && AOSession.RefName != "master" {
			fmt.Printf("\nrefName=%s in AO session file. Consider running command \"ao adm update-ref <refName>\" if this is incorrect\n", AOSession.RefName)
		}
	}
}
