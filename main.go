package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/skatteetaten/ao/pkg/config"

	"github.com/skatteetaten/ao/cmd"
	"github.com/spf13/cobra"
)

const (
	helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:
  {{.UseLine}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if hasSubCommandsAnnotation . "actions"}}

OpenShift Action Commands:{{range .Commands}}{{if eq (index .Annotations "type") "actions"}}
  {{rpad .NameAndAliases .UsagePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if hasSubCommandsAnnotation . "remote"}}

Remote AuroraConfig Commands:{{range .Commands}}{{if eq (index .Annotations "type") "remote"}}
  {{rpad .NameAndAliases .UsagePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if hasSubCommandsAnnotation . "local"}}

Local File Commands:{{range .Commands}}{{if eq (index .Annotations "type") "local"}}
  {{rpad .NameAndAliases .UsagePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (and (eq (index .Annotations "type") "") .IsAvailableCommand)}}
  {{rpad .NameAndAliases .UsagePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

func main() {
	if runtime.GOOS == "windows" {
		if len(os.Args) == 1 {
			if strings.Contains(os.Args[0], "\\") {
				// We presume we have been called from a double click since
				// there are no arguments and the executable contains a catalog name.
				config.Install("", false)
				os.Exit(0)
			}
		}
	}
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
