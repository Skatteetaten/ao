package pingcmd

import "testing"

func TestValidatePingcmd(t *testing.T) {
	var args []string
	var err error

	args = make([]string, 1)
	err = validatePingcmd(args)
	if err != nil {
		t.Errorf("Error in TestValidatePingcmd: %v", err.Error())
	}

	args = make([]string, 0)
	err = validatePingcmd(args)
	if err == nil {
		t.Error("Missing error in TestValidatePingcmd:")
	}

}
