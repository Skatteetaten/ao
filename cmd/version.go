package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the ao client",
	Long:  `Shows the version of the ao client application`,
	Run: func(cmd *cobra.Command, args []string) {

		asJson, _ := cmd.Flags().GetBool("json")

		if !asJson {
			fmt.Println("AO version " + config.Version)
			fmt.Println("Build time " + config.BuildStamp)
			return
		}

		data, err := json.MarshalIndent(config.DefaultAOVersion, "", "  ")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(data))
		}
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("json", "", false, "output version as json")
}
