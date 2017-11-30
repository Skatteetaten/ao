package cmd

import "github.com/spf13/cobra"

var inspectCmd = &cobra.Command{
	Use:    "inspect <deploy-id>",
	Short:  "Inspect a given deploy id",
	Hidden: true,
	RunE:   inspect,
}

func init() {
	RootCmd.AddCommand(inspectCmd)
}

func inspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	result, err := DefaultApiClient.GetApplyResult(args[0])
	if err != nil {
		return err
	}

	cmd.Println(result)

	return nil
}
