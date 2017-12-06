package cmd

import (
	"strings"

	"github.com/pkg/errors"
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

	fileName := args[0]
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}
	path, value := args[1], args[2]

	op := client.JsonPatchOp{
		OP:    "add",
		Path:  path,
		Value: value,
	}

	err := op.Validate()
	if err != nil {
		return err
	}

	res, err := DefaultApiClient.PatchAuroraConfigFile(fileName, op)
	if err != nil {
		return err
	}

	if res != nil {
		return errors.New(res.String())
	}

	cmd.Printf("%s has been updated with %s %s\n", fileName, path, value)

	return nil
}
