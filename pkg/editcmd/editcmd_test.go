package editcmd

import "testing"

func TestStripComments(t *testing.T) {
	const argument = "# Comment\nNo Comment"
	const expected = "No Comment"

	result := stripComments(argument)
	if result != expected {
		t.Errorf("Error in StripComments: Expected %v, got %v", expected, result)
	}
}
