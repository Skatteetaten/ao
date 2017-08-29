package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save changed, new and deleted files for AuroraConfig",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := auroraconfig.Save(config); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Save success")
		}
	},
}

func init() {
	RootCmd.AddCommand(saveCmd)
}
