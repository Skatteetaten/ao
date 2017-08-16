package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:    "gen",
	Short:  "Generates bash completion file",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RootCmd.GenBashCompletionFile("ao.sh"); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("File created")
		}
	},
}

func init() {
	RootCmd.AddCommand(genCmd)
}
