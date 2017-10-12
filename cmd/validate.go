package cmd

import (
	"github.com/spf13/cobra"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

var validateCmd = &cobra.Command{
	Use: "validate",
	Run: func(cmd *cobra.Command, args []string) {
		auroraconfig.Validate(config)
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}
