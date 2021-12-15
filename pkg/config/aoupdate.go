package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/prompt"
)

// Update checks for a new version of ao and performs update with an optional interactive confirmation
// returns true if ao is actually updated
func (aoConfig *AOConfig) Update(noPrompt bool) (bool, error) {
	url, err := aoConfig.getUpdateURL()
	if err != nil {
		return false, err
	}
	logrus.Debugf("Update URL: %s", url)

	serverVersion, err := GetCurrentVersionFromServer(url)
	if err != nil {
		logrus.Warnf("Unable to get ao version from update server on: %s Aborting update detection: %s", url, err)
		return false, nil
	}

	if !serverVersion.IsNewVersion() {
		return false, errors.New("No update available")
	}

	if !noPrompt {
		if runtime.GOOS == "windows" {
			message := fmt.Sprintf("New version of AO is available (%s) - please download from %s", serverVersion.Version, url)
			fmt.Println(message)
			return false, nil
		}
		message := fmt.Sprintf("Do you want to update AO from version %s -> %s?", Version, serverVersion.Version)
		update := prompt.Confirm(message, true)
		if !update {
			return false, errors.New("Update aborted")
		}
	}

	data, err := GetNewAOClient(url)
	if err != nil {
		return false, err
	}

	err = aoConfig.replaceAO(data)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (aoConfig *AOConfig) replaceAO(data []byte) error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	var releasePath string
	// First, we try to write the update to a file in the executable path
	releasePath = executablePath + "_" + "update"
	err = ioutil.WriteFile(releasePath, data, 0755)
	if err != nil {
		// Could not write to executable path, typically because binary is installed in /usr/bin or /usr/local/bin
		// Try the OS Temp Dir
		releasePath = filepath.Join(os.TempDir(), "ao_update")
		err = ioutil.WriteFile(releasePath, data, 0755)
		if err != nil {
			return err
		}
	}
	err = os.Rename(releasePath, executablePath)
	if err != nil {
		err = errors.New("Could not update AO because it is installed in a different file system than temp: " + err.Error())
		return err
	}
	return nil
}
