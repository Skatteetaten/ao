package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Update local repo for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {

		if output, err := auroraconfig.Pull(); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Print(output)
		}
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
