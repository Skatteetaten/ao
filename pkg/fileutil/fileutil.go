package fileutil

import (
	"os"
	"path/filepath"
)

const SpecIllegal = -1
const SpecIsFile = 1
const SpecIsFolder = 2

func IsLegalFileFolder(filespec string) int {
	var err error
	var absolutePath string
	var fi os.FileInfo

	absolutePath, err = filepath.Abs(filespec)
	fi, err = os.Stat(absolutePath)
	if os.IsNotExist(err) {
		return SpecIllegal
	} else {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			return SpecIsFolder
		case mode.IsRegular():
			return SpecIsFile
		}
	}
	return SpecIllegal
}

