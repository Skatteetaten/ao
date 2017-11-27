package cmd

import (
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:         "set <file> <json-path> <value>",
	Short:       "Sets the config to the value for the given AuroraConfig file",
	Annotations: map[string]string{"type": "remote"},
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

	cmd.Printf("%s has been updated\n", fileName)

	return nil
}
