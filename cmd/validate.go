package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var flagFullValidation bool

var validateCmd = &cobra.Command{
	Use:         "validate",
	Short:       "Validate local modifications in the current AuroraConfig",
	Annotations: map[string]string{"type": "local"},
	RunE:        Validate,
}

func init() {
	RootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&flagAffiliation, "auroraconfig", "a", "", "AuroraConfig to validate")
	validateCmd.Flags().BoolVarP(&flagFullValidation, "full", "f", false, "Validate resources")
}

func Validate(cmd *cobra.Command, args []string) error {
	wd, _ := os.Getwd()
	gitRoot, err := versioncontrol.FindGitPath(wd)
	if err != nil {
		return err
	}

	if flagAffiliation != "" {
		DefaultApiClient.Affiliation = flagAffiliation
	}

	ac, err := versioncontrol.CollectAuroraConfigFilesInRepo(DefaultApiClient.Affiliation, gitRoot)
	if err != nil {
		return err
	}

	res, err := DefaultApiClient.ValidateAuroraConfig(ac, flagFullValidation)
	if err != nil {
		return err
	}
	if res != nil {
		return errors.New(res.String())
	}
	fmt.Println("OK")

	return nil
}
