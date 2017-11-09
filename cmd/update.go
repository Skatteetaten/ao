package cmd

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/command"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates for the ao client, and downloads the update if available.",
	Long:  `Available updates are searched for using a service in the OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		serverVersion, err := ao.GetCurrentVersionFromServer()
		if err != nil {
			fmt.Println(err)
			return
		}

		if !serverVersion.IsNewVersion(config.Version) {
			fmt.Println("No update available")
			return
		}

		message := fmt.Sprintf("Do you want update to version %s?", serverVersion.Version)
		update := prompt.Confirm(message)
		if !update {
			return
		}

		data, err := ao.GetNewAOClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		command.ReplaceAO(data)
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
