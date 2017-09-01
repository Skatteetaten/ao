package cmd

import (
	"fmt"

	_ "go/token"
	"os"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Checkout AuroraConfig (git repository) for current affiliation",
	Run: func(cmd *cobra.Command, args []string) {
		affiliation := config.GetAffiliation()

		if affiliationFlag, _ := cmd.LocalFlags().GetString("affiliation"); len(affiliationFlag) > 0 {
			affiliation = affiliationFlag
		}

		path, _ := cmd.LocalFlags().GetString("path")

		if len(path) < 1 {
			wd, _ := os.Getwd()
			path = fmt.Sprintf("%s/%s", wd, affiliation)
		}

		url := getGitUrl(affiliation, userName)

		fmt.Printf("Cloning AuroraConfig for affiliation %s\n", affiliation)
		fmt.Printf("From: %s\n\n", url)

		if output, err := auroraconfig.Checkout(url, path); err != nil {
			return
		} else {
			fmt.Print(output)
		}

		if err := aoConfig.AddCheckoutPath(affiliation, path, aoConfigLocation); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Checkout success")
	},
}

func getGitUrl(affiliation, user string) string {
	clientConfig, err := auroraconfig.GetClientConfig(config)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	gitUrlPattern := clientConfig.GitUrlPattern

	if !strings.Contains(gitUrlPattern, "https://") {
		return fmt.Sprintf(gitUrlPattern, affiliation)
	}

	host := strings.TrimPrefix(gitUrlPattern, "https://")
	newPattern := fmt.Sprintf("https://%s@%s", user, host)
	return fmt.Sprintf(newPattern, affiliation)
}

func init() {
	RootCmd.AddCommand(checkoutCmd)

	viper.BindEnv("USER")
	checkoutCmd.Flags().StringP("affiliation", "a", "", "Affiliation to clone")
	checkoutCmd.Flags().StringP("path", "p", "", "Checkout repo to path")
	checkoutCmd.Flags().StringP("user", "u", viper.GetString("USER"), "Checkout repo as user")
}
