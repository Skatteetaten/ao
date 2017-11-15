package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
	_ "go/token"
	"os"
)

var (
	flagCheckoutAffiliation string
	flagCheckoutPath        string
	flagCheckoutUser        string
)

var checkoutCmd = &cobra.Command{
	Use:         "checkout",
	Short:       "Checkout AuroraConfig (git repository) for current affiliation",
	Annotations: map[string]string{"type": "local"},
	RunE:        Checkout,
}

func init() {
	RootCmd.AddCommand(checkoutCmd)

	user, _ := os.LookupEnv("USER")
	checkoutCmd.Flags().StringVarP(&flagCheckoutAffiliation, "affiliation", "a", "", "Affiliation to clone")
	checkoutCmd.Flags().StringVarP(&flagCheckoutPath, "path", "p", "", "Checkout repo to path")
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

	url := versioncontrol.GetGitUrl(affiliation, flagCheckoutUser, DefaultApiClient)

	logrus.Debug(url)
	fmt.Printf("Cloning AuroraConfig for affiliation %s\n", affiliation)
	fmt.Printf("From: %s\n\n", url)

	output, err := versioncontrol.Checkout(url, path)
	if err != nil {
		return err
	} else {
		fmt.Print(output)
	}

	fmt.Println("Checkout success")
	return nil
}
