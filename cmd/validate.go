package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:         "validate",
	Short:       "Validate local modifications in the current AuroraConfig",
	Annotations: map[string]string{"type": "local"},
	RunE:        Validate,
}

func init() {
	RootCmd.AddCommand(validateCmd)
}

func Validate(cmd *cobra.Command, args []string) error {
	ac, err := versioncontrol.CollectFiles()
	if err != nil {
		return err
	}

	res, err := DefaultApiClient.ValidateAuroraConfig(ac)
	if err != nil {
		return err
	}
	if res != nil {
		return errors.New(res.String())
	}
	fmt.Println("OK")

	return nil
}
