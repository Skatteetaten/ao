package executil

import "testing"

func TestRunInteractively(t *testing.T) {
	err := RunInteractively("foobar-command", "foobar-folder", "Foo", "Bar")
	if err == nil {
		t.Error("Error in TestRunInteractively: Expected error from foobar run")
	}
}
