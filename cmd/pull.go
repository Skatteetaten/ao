package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Update local repo for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {

		if output, err := versioncontrol.Pull(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Print(output)
		}
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
