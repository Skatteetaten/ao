package cmdoptions

import (
	"strings"
	"testing"
)

func TestListOptions(t *testing.T) {
	var opt CommonCommandOptions

	opt.Verbose = true
	opt.DryRun = true

	output := opt.ListOptions()

	const expectedVerbose = "Verbose: true"
	if !strings.Contains(output, expectedVerbose) {
		t.Errorf("Expected %v, got %v", expectedVerbose, output)
	}

	const expectedLocalhost = "Localhost: false"
	if !strings.Contains(output, expectedLocalhost) {
		t.Errorf("Expected %v, got %v", expectedLocalhost, output)
	}
}
