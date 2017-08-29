package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "go/token"
	"os"
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

		url := ""
		if gitUrlPattern, err := auroraconfig.GetGitLocation(config); err != nil {
			fmt.Println(err)
			return
		} else {
			url = fmt.Sprintf(gitUrlPattern, affiliation)
		}

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

func init() {
	RootCmd.AddCommand(checkoutCmd)

	viper.BindEnv("USER")
	checkoutCmd.Flags().StringP("affiliation", "a", "", "Affiliation to clone")
	checkoutCmd.Flags().StringP("path", "p", "", "Checkout repo to path")
	checkoutCmd.Flags().StringP("user", "u", viper.GetString("USER"), "Checkout repo as user")
}
