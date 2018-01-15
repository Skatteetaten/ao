package cmd

import (
	"runtime"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:         "pull",
	Short:       "Update local repo for AuroraConfig",
	Annotations: map[string]string{"type": "local"},
	RunE:        Pull,
}

func init() {
	if runtime.GOOS != "windows" {
		RootCmd.AddCommand(pullCmd)
	}
}

func Pull(cmd *cobra.Command, args []string) error {

	if output, err := versioncontrol.Pull(); err != nil {
		return err
	} else {
		cmd.Print(output)
	}

	return nil
}
