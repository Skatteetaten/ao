package cmd

import (
	"fmt"
	"os"

	"github.com/skatteetaten/ao/pkg/client"

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
	validateCmd.Flags().StringVarP(&flagAffiliation, "auroraconfig", "a", "", "AuroraConfig to validate")
}

func Validate(cmd *cobra.Command, args []string) error {
	wd, _ := os.Getwd()
	gitRoot, err := versioncontrol.FindGitPath(wd)
	if err != nil {
		return err
	}

	files, err := versioncontrol.CollectJSONFilesInRepo(gitRoot)
	if err != nil {
		return err
	}

	if flagAffiliation != "" {
		DefaultApiClient.Affiliation = flagAffiliation
	}

	ac := &client.AuroraConfig{
		Files:    files,
		Versions: make(map[string]string),
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
