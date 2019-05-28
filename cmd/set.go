package cmd

import (
	"github.com/spf13/cobra"

	"github.com/skatteetaten/ao/pkg/service"
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

	name, path, value := args[0], args[1], args[2]

	fileName, err := service.SetValue(DefaultApiClient, name, path, value)
	if err != nil {
		return err
	}

	cmd.Printf("%s has been updated with %s %s\n", fileName, path, value)

	return nil
}
