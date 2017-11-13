package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Update local repo for AuroraConfig",
	RunE:  Pull,
}

func init() {
	RootCmd.AddCommand(pullCmd)
}

func Pull(cmd *cobra.Command, args []string) error {

	if output, err := versioncontrol.Pull(); err != nil {
		return err
	} else {
		fmt.Print(output)
	}

	return nil
}
