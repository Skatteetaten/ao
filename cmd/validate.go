package cmd

import (
	"os"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
)

var flagFullValidation bool
var flagRemoteValidation bool

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
	validateCmd.Flags().BoolVarP(&flagRemoteValidation, "remote", "r", false, "Validate remote AuroraConfig instead of local files")
}

func Validate(cmd *cobra.Command, args []string) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	gitRoot, err := versioncontrol.FindGitPath(wd)
	if err != nil {
		return err
	}

	if flagAuroraConfig != "" {
		DefaultApiClient.Affiliation = flagAuroraConfig
	}

	var warnings string
	if flagRemoteValidation {
		cmd.Printf("Validating remote AuroraConfig=%s@%s fullValidation=%t\n", DefaultApiClient.Affiliation, DefaultApiClient.RefName, flagFullValidation)
		warnings, err = DefaultApiClient.ValidateRemoteAuroraConfig(flagFullValidation)
		if err != nil {
			return err
		}
	} else {
		ac, err := versioncontrol.CollectAuroraConfigFilesInRepo(DefaultApiClient.Affiliation, gitRoot)
		if err != nil {
			return err
		}
		cmd.Printf("Validating AuroraConfig=%s gitRoot=%s fullValidation=%t\n", DefaultApiClient.Affiliation, gitRoot, flagFullValidation)
		warnings, err = DefaultApiClient.ValidateAuroraConfig(ac, flagFullValidation)
		if err != nil {
			return err
		}
	}

	if warnings != "" {
		cmd.Println("")
		cmd.Println("AuroraConfig contains the following warnings:")
		cmd.Println("")
		cmd.Println(warnings)
	} else {
		cmd.Println("OK")
	}

	return nil
}
