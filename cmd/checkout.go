package cmd

import (
	"fmt"

	_ "go/token"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

// TODO: Change affiliation to auroraconfig, flags
var (
	flagCheckoutAffiliation string
	flagCheckoutPath        string
	flagCheckoutUser        string
)

var checkoutCmd = &cobra.Command{
	Use:         "checkout",
	Short:       "Checkout the given AuroraConfig (git repository)",
	Annotations: map[string]string{"type": "local"},
	RunE:        Checkout,
}

func init() {
	RootCmd.AddCommand(checkoutCmd)

	user, _ := os.LookupEnv("USER")
	checkoutCmd.Flags().StringVarP(&flagCheckoutAffiliation, "auroraconfig", "a", "", "AuroraConfig to clone")
	checkoutCmd.Flags().StringVarP(&flagCheckoutPath, "path", "", "", "Checkout repo to path")
	checkoutCmd.Flags().StringVarP(&flagCheckoutUser, "user", "u", user, "Checkout repo as user")
}

func Checkout(cmd *cobra.Command, args []string) error {
	affiliation := AO.Affiliation
	if flagCheckoutAffiliation != "" {
		affiliation = flagCheckoutAffiliation
	}

	wd, _ := os.Getwd()
	path := fmt.Sprintf("%s/%s", wd, affiliation)
	if flagCheckoutPath != "" {
		path = flagCheckoutPath
	}

	clientConfig, err := DefaultApiClient.GetClientConfig()
	if err != nil {
		return err
	}
	url := versioncontrol.GetGitUrl(affiliation, flagCheckoutUser, clientConfig.GitUrlPattern)

	logrus.Debug(url)
	fmt.Printf("Cloning AuroraConfig %s\n", affiliation)
	fmt.Printf("From: %s\n\n", url)

	if err := versioncontrol.Checkout(url, path); err != nil {
		return err
	}

	if err := versioncontrol.CreatePreCommitHook(path, affiliation); err != nil {
		return err
	}

	fmt.Println("Checkout success")
	return nil
}
