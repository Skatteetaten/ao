package printutil

import (
	"testing"
)

func TestFormatTable(t *testing.T) {
	headers := []string{"Foo", "Bar"}
	column1 := []string{"POTUS", "Statsminister"}
	column2 := []string{"FLOUTUS"}

	output := FormatTable(headers, column1, column2)
	outputLen := len(output)
	if outputLen != 67 {
		t.Error("Illegal length")
	}

}
