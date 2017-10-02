package deletecmd

import (
	"testing"
)

func TestAddDeleteFile(t *testing.T) {
	var deletecmd *DeletecmdClass
	deletecmd = new(DeletecmdClass)

	const argument = "foobar"
	const expected = "foobar"
	deletecmd.addDeleteFile(argument)
	if len(deletecmd.deleteFileList) != 1 {
		t.Errorf("Error in TestAddDeleteFile: Len=%v", len(deletecmd.deleteFileList))
	} else {
		output := deletecmd.deleteFileList[0]

		if output != expected {
			t.Errorf("Error in TestAddDeleteFile: Expected %v, got %v", expected, output)
		}
	}
}
