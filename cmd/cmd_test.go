package cmd

import (
	"testing"
)

func TestInitConfigCobra(t *testing.T) {
	initConfigCobra()
	if aoConfig == nil {
		t.Error("InitConfigCobra did not create a config")
	}
}
