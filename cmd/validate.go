package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate local AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {
		ac, err := versioncontrol.CollectFiles()
		if err != nil {
			fmt.Println(err)
			return
		}

		res, err := DefaultApiClient.ValidateAuroraConfig(ac)
		if err != nil {
			fmt.Println(err)
			return
		}
		if res != nil {
			res.PrintAllErrors()
		} else {
			fmt.Println("OK")
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}
