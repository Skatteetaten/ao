package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"os"
)

var completionCmd = &cobra.Command{
	Use:    "completion",
	Short:  "Generates bash completion file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := RootCmd.GenBashCompletionFile("ao.sh"); err == nil {
			wd, _ := os.Getwd()
			fmt.Println("Bash completion file created at", wd + "/ao.sh")
		} else {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
