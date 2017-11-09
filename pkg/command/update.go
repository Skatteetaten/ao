package command

import (
	"io/ioutil"
	"os"
)

func ReplaceAO(data []byte) error {

	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	releasePath := executablePath + "_" + "update"
	err = ioutil.WriteFile(releasePath, data, 0750)
	if err != nil {
		return err
	}
	err = os.Rename(releasePath, executablePath)
	if err != nil {
		return err
	}

	return nil
}
