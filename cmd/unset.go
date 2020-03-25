package cmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
)

const unsetExample = `  ao unset foo.json /pause

  ao unset test/foo.json /config/IMPORTANT_ENV`

var unsetCmd = &cobra.Command{
	Use:         "unset <file> <json-path>",
	Short:       "Remove a single configuration entry in the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     unsetExample,
	RunE:        Unset,
}

const unsetExampleNewExperimental = `  ao unsetnew foo.json /pause

  ao unsetnew test/foo.json /config/IMPORTANT_ENV`

var unsetNewCmd = &cobra.Command{
	Use:         "unsetnew <file> <json-path>",
	Short:       "Remove a single configuration entry in the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     unsetExampleNewExperimental,
	RunE:        UnsetNew,
}

func init() {
	RootCmd.AddCommand(unsetCmd)
	RootCmd.AddCommand(unsetNewCmd)
}

func Unset(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	name := args[0]
	fileName, err := fileNames.Find(name)
	if err != nil {
		return err
	}

	path := args[1]
	op := auroraconfig.JsonPatchOp{
		OP:   "remove",
		Path: path,
	}

	if err = op.Validate(); err != nil {
		return err
	}

	if err = DefaultApiClient.PatchAuroraConfigFile(fileName, op); err != nil {
		return err
	}

	cmd.Printf("%s has been updated\n", fileName)

	return nil
}

func UnsetNew(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Usage()
	}

	fileNames, err := DefaultApiClient.GetFileNames()
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
	auroraConfigFile, eTag, err := DefaultApiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	// Remove entry at path
	if err := auroraconfig.RemoveEntry(auroraConfigFile, path); err != nil {
		return err
	}

	// Save config file
	if err := DefaultApiClient.PutAuroraConfigFile(auroraConfigFile, eTag); err != nil {
		return err
	}

	cmd.Printf("%s has been updated\n", fileName)

	return nil
}
