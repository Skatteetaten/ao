package cmd

import (
	"os"

	"ao/pkg/versioncontrol"
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

// Validate is the entry point of the `validate` cli command
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
		DefaultAPIClient.Affiliation = flagAuroraConfig
	}

	var warnings string
	if flagRemoteValidation {
		cmd.Printf("Validating remote AuroraConfig=%s@%s fullValidation=%t\n", DefaultAPIClient.Affiliation, DefaultAPIClient.RefName, flagFullValidation)
		warnings, err = DefaultAPIClient.ValidateRemoteAuroraConfig(flagFullValidation)
		if err != nil {
			return err
		}
	} else {
		ac, err := versioncontrol.CollectAuroraConfigFilesInRepo(DefaultAPIClient.Affiliation, gitRoot)
		if err != nil {
			return err
		}
		cmd.Printf("Validating AuroraConfig=%s gitRoot=%s fullValidation=%t\n", DefaultAPIClient.Affiliation, gitRoot, flagFullValidation)
		warnings, err = DefaultAPIClient.ValidateAuroraConfig(ac, flagFullValidation)
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
