package updatecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/versionutil"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const ao5Url = "http://ao-update-service-paas-ao-update.utv.paas.skead.no"

func InstallAuroraOpenshiftGenerator() (err error) {
	const gitExec = "git"
	const gitCloneCommand = "clone"
	const generatorUrl = "https://github.com/Skatteetaten/generator-aurora-openshift.git"
	const generatorFolderName = "generator-aurora-openshift"
	const generatorInstallFile = "install.sh"
	const shellExecutable = "sh"

	// Make a temp folder
	installDir, err := ioutil.TempDir("", "aog-install")
	if err != nil {
		return err
	}

	// Clone the generator
	err = executil.RunInteractively(gitExec, installDir, gitCloneCommand, generatorUrl)
	if err != nil {
		return err
	}

	// Run install
	err = executil.RunInteractively(shellExecutable, filepath.Join(installDir, generatorFolderName), generatorInstallFile)
	if err != nil {
		return err
	}

	// Remove temp folder
	err = os.RemoveAll(installDir)
	if err != nil {
		return err
	}

	return
}

func UpdateSelf(args []string, simulate bool, forceVersion string, forceUpdate bool) (output string, err error) {
	var releaseVersion string

	if forceVersion == "" {
		releaseVersion, err = getReleaseVersion()
		if err != nil {
			return "", errors.New("Update server unreachable: " + err.Error())
		}
	} else {
		releaseVersion = forceVersion
	}

	myVersion, _ := getMyVersion()

	if myVersion != releaseVersion || forceUpdate {
		output += "New version detected: Current version: " + myVersion + ".  Available version: " + releaseVersion
		if !simulate {
			err = doUpdate(releaseVersion)
			if err != nil {
				return
			}
			output += "\nAO updated sucessfully"
		}
	} else {
		output += "No update available"
	}

	return
}

func doUpdate(version string) (err error) {
	releaseFilename := "ao_" + version
	releaseUrl := ao5Url + "/" + releaseFilename

	executablePath, err := os.Executable()
	if err != nil {
		return
	}

	releasePath := executablePath + "_" + version
	body, err := getFile(releaseUrl)
	err = ioutil.WriteFile(releasePath, []byte(body), 0750)
	if err != nil {
		return
	}
	err = os.Rename(releasePath, executablePath)
	if err != nil {
		return
	}
	return
}

func getReleaseVersion() (version string, err error) {
	releaseinfoUrl := ao5Url + "/releaseinfo.json"
	releaseinfo, err := getFile(releaseinfoUrl)
	if err != nil {
		return
	}
	releaseVersionStruct, err := versionutil.Json2Version(releaseinfo)

	version = releaseVersionStruct.Version
	return
}

func getMyVersion() (version string, err error) {
	var versionStruct versionutil.VersionStruct
	versionStruct.Init()

	version = versionStruct.Version
	return
}

func getFile(url string) (file []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {

		err = errors.New(fmt.Sprintf("Error downloading update from %v: %v", ao5Url, err))
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
		return
	}

	file, err = ioutil.ReadAll(resp.Body)
	return
}
