package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates for the ao client, and downloads the update if available.",
	Long:  `Available updates are searched for using a service in the OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ao.Update()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("AO has been updated")
		}
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
