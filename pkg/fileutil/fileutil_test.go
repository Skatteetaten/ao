package fileutil

import (
	"runtime"
	"testing"
)

func TestIsLegalFileFolder(t *testing.T) {
	var expected, res int
	var testfolder string

	expected = SpecIsFolder
	if runtime.GOOS == "windows" {
		testfolder = "C:\\Windows"
	} else {
		testfolder = "/bin"
	}
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize bin/windows as folder")
	}

	expected = SpecIsFile
	if runtime.GOOS == "windows" {
		testfolder = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	} else {
		testfolder = "/etc/hosts"
	}
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize hosts file as legal file")
	}

	expected = SpecIllegal
	testfolder = "/Go/is/an/open/source/programming/language/that/makes/it/easy/to/build/" +
		"simple/reliable/and/efficient/software."
	res = IsLegalFileFolder(testfolder)
	if res != expected {
		t.Error("Failed to recognize illegal folder")
	}
}
