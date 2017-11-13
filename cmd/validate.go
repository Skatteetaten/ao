package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate local AuroraConfig",
	RunE:  Validate,
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
	} else {
		fmt.Println("OK")
	}

	return nil
}
