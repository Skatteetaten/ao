package importcmd

import "testing"

func TestValidateImportCommand(t *testing.T) {
	var args []string
	args = make([]string, 0)
	err := validateImportCommand(args)
	if err == nil {
		t.Error("Error in TestValidateImportCommand: Expected error message")
	}

}
