package cmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
)

const setExample = `  ao set foo.json /pause true

  ao set test/about.json /cluster utv

  ao set test/foo.yaml /config/IMPORTANT_ENV 'Hello World'`

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
	fileName, path, value := args[0], args[1], args[2]

	// Load config file
	auroraConfigFile, eTag, err := DefaultApiClient.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	// Set value
	if err := auroraconfig.SetValue(auroraConfigFile, path, value); err != nil {
		return err
	}

	// Save config file
	if err := DefaultApiClient.PutAuroraConfigFile(auroraConfigFile, eTag); err != nil {
		return err
	}

	cmd.Printf("%s has been updated with %s %s\n", fileName, path, value)

	return nil
}
