package cmd

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/spf13/cobra"
)

const unsetExample = `  ao unset foo.json /pause

  ao unset test/foo.json /config/IMPORTANT_ENV`

var unsetCmd = &cobra.Command{
	Use:         "unset <file> <json-path>",
	Short:       "Remove a single configuration value in the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     unsetExample,
	RunE:        Unset,
}

func init() {
	RootCmd.AddCommand(unsetCmd)
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
	op := client.JsonPatchOp{
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
