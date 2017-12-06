package cmd

import (
	"strings"

	"github.com/pkg/errors"
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

	fileName := args[0]
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}
	path := args[1]

	op := client.JsonPatchOp{
		OP:   "remove",
		Path: path,
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

	cmd.Printf("%s has been updated\n", fileName)

	return nil
}
