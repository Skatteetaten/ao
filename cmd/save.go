package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/versioncontrol"
	"github.com/spf13/cobra"
	"os"
)

var flagSaveAsUser string

var saveCmd = &cobra.Command{
	Use:         "save",
	Short:       "Save changed, new and deleted files for AuroraConfig",
	Annotations: map[string]string{"type": "local"},
	RunE:        Save,
}

func init() {
	RootCmd.AddCommand(saveCmd)

	user, _ := os.LookupEnv("USER")
	saveCmd.Flags().StringVarP(&flagSaveAsUser, "user", "u", user, "Save AuroraConfig as user")
}

func Save(cmd *cobra.Command, args []string) error {
	url := versioncontrol.GetGitUrl(AO.Affiliation, flagSaveAsUser, DefaultApiClient)

	if _, err := versioncontrol.Save(url, DefaultApiClient); err != nil {
		return err
	} else {
		fmt.Println("Save success")
	}

	return nil
}
