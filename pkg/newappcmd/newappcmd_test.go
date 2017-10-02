package newappcmd

import "testing"

func TestReadGeneratorValues(t *testing.T) {
	_, err := readGeneratorValues("foobar")
	if err == nil {
		t.Error("Error in TestReadGeneratorValues: Expected error on foobar run")
	}
}
