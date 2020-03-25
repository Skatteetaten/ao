package cmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/service"
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

const setNewExample = `  ao setnew foo.json /pause true

  ao setnew test/about.json /cluster utv

  ao setnew test/foo.json /config/IMPORTANT_ENV 'Hello World'`

var setCmdNewExperimental = &cobra.Command{
	Use:         "setnew <file> <json-path> <value>",
	Short:       "Set a single configuration value in the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     setNewExample,
	RunE:        SetNew,
}

func init() {
	RootCmd.AddCommand(setCmd)
	RootCmd.AddCommand(setCmdNewExperimental)
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

func SetNew(cmd *cobra.Command, args []string) error {
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
