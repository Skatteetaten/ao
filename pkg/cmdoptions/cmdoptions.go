package cmdoptions

import (
	"fmt"
)

type CommonCommandOptions struct {
	Verbose     bool
	Debug       bool
	DryRun      bool
	Localhost   bool
	ShowConfig  bool
	Pretty      bool
	ShowObjects bool
	ServerApi   string
	Token       string
}

func (opt *CommonCommandOptions) ListOptions() (output string) {
	output = fmt.Sprintf("Verbose: %v, Debug: %v, DryRun %v, Localhost: %v, ShowConfig: %v, ShowObject: %v, ServerApi: %v",
		opt.Verbose, opt.Debug, opt.DryRun, opt.Localhost, opt.ShowConfig, opt.ShowObjects, opt.ServerApi)
	return
}
