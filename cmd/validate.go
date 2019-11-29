package cmd

import (
	"fmt"
	"os"

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
	validateCmd.Flags().StringVarP(&flagAuroraConfig, "auroraconfig", "a", "", "AuroraConfig to validate")
	validateCmd.Flags().BoolVarP(&flagFullValidation, "full", "f", false, "Validate resources")
}

func Validate(cmd *cobra.Command, args []string) error {
	wd, _ := os.Getwd()
	gitRoot, err := versioncontrol.FindGitPath(wd)
	if err != nil {
		return err
	}

	if flagAuroraConfig != "" {
		DefaultApiClient.Affiliation = flagAuroraConfig
	}

	ac, err := versioncontrol.CollectAuroraConfigFilesInRepo(DefaultApiClient.Affiliation, gitRoot)
	if err != nil {
		return err
	}


	fmt.Printf("Validating auroraAonfig=%s gitRoot=%s fullValidation=%t\n", DefaultApiClient.Affiliation, gitRoot, flagFullValidation)

	if err := DefaultApiClient.ValidateAuroraConfig(ac, flagFullValidation); err != nil {
		return err
	}

	fmt.Println("OK")

	return nil
}
