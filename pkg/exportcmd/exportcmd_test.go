package exportcmd

import "testing"

func TestGetAdc(t *testing.T) {
	var exportObj *ExportcmdClass
	exportObj = new(ExportcmdClass)

	output, _ := exportObj.getAdc(nil)
	if output != notYetImplemented {
		t.Errorf("Error in TestGetAdc: Expected %v, got %v", notYetImplemented, output)
	}
}
