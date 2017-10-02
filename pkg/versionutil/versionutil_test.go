package versionutil

import "testing"

func TestVersion2Text(t *testing.T) {
	var version *VersionStruct
	version = new(VersionStruct)

	version.Version = "1.0.0"
	var expected string = "Aurora Oc version " + version.Version
	output, _ := version.Version2Text()
	if output != expected {
		t.Errorf("Error in TestVersion2Text: Expected %v, got %v", expected, output)
	}
}
