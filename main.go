package main

import (
	"fmt"
	"github.com/skatteetaten/ao/cmd"
	"github.com/spf13/cobra"
	"os"
)

const (
	helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.CommandPath}} [command] [flags]{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if hasSubCommandsAnnotation . "actions"}}

OpenShift Action Commands:{{range .Commands}}{{if eq (index .Annotations "type") "actions"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if hasSubCommandsAnnotation . "remote"}}

Remote AuroraConfig Commands:{{range .Commands}}{{if eq (index .Annotations "type") "remote"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if hasSubCommandsAnnotation . "local"}}

Local File Commands:{{range .Commands}}{{if eq (index .Annotations "type") "local"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (and (eq (index .Annotations "type") "") (ne .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

func main() {
	cmd.RootCmd.SetHelpTemplate(helpTemplate)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.AddTemplateFunc("hasSubCommandsAnnotation", func(cmd cobra.Command, annotation string) bool {
		for _, c := range cmd.Commands() {
			t := c.Annotations["type"]
			if t == annotation {
				return true
			}
		}

		return false
	})
}
