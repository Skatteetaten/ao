package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates for the ao client, and downloads the update if available.",
	Long:  `Available updates are searched for using a service in the OpenShift cluster.`,
	RunE:  Update,
}

func init() {
	if runtime.GOOS != "windows" {
		RootCmd.AddCommand(updateCmd)
	}
}

// Update is the entry point of the `update` cli command
func Update(cmd *cobra.Command, args []string) error {
	updated, err := AOConfig.Update(true)
	if err != nil {
		return err
	}
	if updated {
		fmt.Println("AO has been updated")
	}

	return nil
}
