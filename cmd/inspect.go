package cmd

import "github.com/spf13/cobra"
import "strings"
import "github.com/pkg/errors"

var inspectCmd = &cobra.Command{
	Use:         "inspect <deploy-id>",
	Short:       "Inspect a given deploy id",
	Annotations: map[string]string{"type": "remote"},
	RunE:        inspect,
}

func init() {
	RootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringVarP(&flagAuroraConfig, "auroraconfig", "a", "", "set auroraconfigId to which the deploy-id belongs to")
}

func inspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	if flagAuroraConfig != "" {
		DefaultApiClient.Affiliation = flagAuroraConfig
	}
	result, err := DefaultApiClient.GetApplyResult(args[0])
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			return errors.Errorf("could not find deploy-id %s for AuroraConfig %s", args[0], DefaultApiClient.Affiliation)
		}
		return err
	}

	cmd.Println(result)

	return nil
}
