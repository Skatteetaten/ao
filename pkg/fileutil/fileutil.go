package fileutil

import (
	"errors"
	"fmt"
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

func ValidateFileFolderArg(args []string) (error error) {
	var errorString string

	if len(args) == 0 {
		errorString += "Missing file/folder "
	} else {
		// Chceck argument 0 for legal file / folder
		validateCode := IsLegalFileFolder(args[0])
		if validateCode < 0 {
			errorString += fmt.Sprintf("Illegal file / folder: %v\n", args[0])
		}

	}

	if errorString != "" {
		return errors.New(errorString)
	}
	return
}
