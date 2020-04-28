package cmd

import (
	"fmt"
	"runtime"

	// blank import that is needed for side effects at initialization
	_ "go/token"
	"os"

	"ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

// TODO: Change affiliation to auroraconfig, flags
var (
	flagCheckoutAffiliation string
	flagCheckoutPath        string
	flagCheckoutUser        string
	flagGitHookType         string
)

var checkoutCmd = &cobra.Command{
	Use:         "checkout",
	Short:       "Checkout the given AuroraConfig (git repository)",
	Annotations: map[string]string{"type": "local"},
	RunE:        Checkout,
}

func init() {
	if runtime.GOOS != "windows" {
		RootCmd.AddCommand(checkoutCmd)

		user, _ := os.LookupEnv("USER")
		checkoutCmd.Flags().StringVarP(&flagCheckoutAffiliation, "auroraconfig", "a", "", "AuroraConfig to clone")
		checkoutCmd.Flags().StringVarP(&flagCheckoutPath, "path", "", "", "Checkout repo to path")
		checkoutCmd.Flags().StringVarP(&flagCheckoutUser, "user", "u", user, "Checkout repo as user")
	}
}

// Checkout is the main method for the `checkout` cli command
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

	clientConfig, err := DefaultAPIClient.GetClientConfig()
	if err != nil {
		return err
	}
	url := versioncontrol.GetGitUrl(affiliation, flagCheckoutUser, clientConfig.GitUrlPattern)

	fmt.Printf("Cloning AuroraConfig %s\n", affiliation)
	fmt.Printf("From: %s\n\n", url)

	if err := versioncontrol.Checkout(url, path); err != nil {
		return err
	}

	if err := versioncontrol.CreateGitValidateHook(path, flagGitHookType, affiliation); err != nil {
		return err
	}

	fmt.Println("Checkout success")
	return nil
}
