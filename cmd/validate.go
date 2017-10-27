package cmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"fmt"
)

var validateCmd = &cobra.Command{
	Use: "validate",
	Run: func(cmd *cobra.Command, args []string) {
		mainMessage, messages, err := auroraconfig.Validate(config)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(mainMessage)
		for _, m := range messages {
			fmt.Println(m)
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}
