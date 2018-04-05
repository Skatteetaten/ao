package cmd

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/spf13/cobra"
)

const setExample = `  ao set foo.json /pause true

  ao set test/about.json /cluster utv

  ao set test/foo.json /config/IMPORTANT_ENV 'Hello World'`

var setCmd = &cobra.Command{
	Use:         "set <file> <json-path> <value>",
	Short:       "Set a single configuration value in the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     setExample,
	RunE:        Set,
}

func init() {
	RootCmd.AddCommand(setCmd)
}

func Set(cmd *cobra.Command, args []string) error {
	if len(args) != 3 {
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

	path, value := args[1], args[2]
	op := client.JsonPatchOp{
		OP:    "add",
		Path:  path,
		Value: value,
	}

	if err = op.Validate(); err != nil {
		return err
	}

	if err = DefaultApiClient.PatchAuroraConfigFile(fileName, op); err != nil {
		return err
	}

	cmd.Printf("%s has been updated with %s %s\n", fileName, path, value)

	return nil
}
