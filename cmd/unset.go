package cmd

import (
	"ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
)

const unsetExample = `  ao unset foo.json /pause

  ao unset test/foo.json /config/IMPORTANT_ENV

  ao unset test/bar.yaml /config/DEBUG`

var unsetCmd = &cobra.Command{
	Use:         "unset <file> <path-to-key>",
	Short:       "Remove a single configuration key from the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     unsetExample,
	RunE:        Unset,
}

func init() {
	RootCmd.AddCommand(unsetCmd)
}

// Unset is the entry point of the `unset` cli command
func Unset(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	fileNames, err := DefaultAPIClient.GetFileNames()
	if err != nil {
		return err
	}

	name := args[0]
	fileName, err := fileNames.Find(name)
	if err != nil {
		return err
	}

	path := args[1]

	// Load config file
	auroraConfigFile, eTag, err := DefaultAPIClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	// Remove entry at path
	if err := auroraconfig.RemoveEntry(auroraConfigFile, path); err != nil {
		return err
	}

	// Save config file
	if err := DefaultAPIClient.PutAuroraConfigFile(auroraConfigFile, eTag); err != nil {
		return err
	}

	cmd.Printf("%s has been updated\n", fileName)

	return nil
}
