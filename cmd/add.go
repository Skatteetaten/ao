package cmd

import (
	"github.com/spf13/cobra"
)

const addExample = `  Given the following AuroraConfig:
    - about.json
    - foobar.json
    - bar.json
    - foo/about.json
    - foo/bar.json
    - foo/foobar.json

  # adds test/about.json to AuroraConfig
  ao add test/about.json ./about.json

  # will throw an error because about.json already exists
  ao add about.json ./about.json

  # adds prod/about to AuroraConfig
  ao add prod/about ~/files/about.json`

var addCmd = &cobra.Command{
	Use:         "add <name> <file>",
	Short:       "Add a single file to the current AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	Example:     addExample,
	RunE:        AddGraphql,
}

func init() {
	RootCmd.AddCommand(addCmd)
}
