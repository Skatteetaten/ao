package updatecmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/versionutil"
)

const updateServerPattern = "http://ao-aurora-tools.%s.paas.skead.no"

const aoDownloadPath = "/api/ao"
const getCurrentVersionPath = "/api/version"

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

func UpdateSelf(args []string, simulate bool, forceVersion string, forceUpdate bool, config *configuration.ConfigurationClass) (output string, err error) {
	var releaseVersion string
	var url string

	cluster := config.GetApiClusterName()
	if cluster == "utv" {
		url = fmt.Sprintf(updateServerPattern, "utv")
	} else {
		url = fmt.Sprintf(updateServerPattern, "test")
	}

	if forceVersion == "" {
		releaseVersion, err = getReleaseVersion(url)
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
			err = doUpdate(url, releaseVersion)
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

func doUpdate(url string, version string) (err error) {

	executablePath, err := os.Executable()
	if err != nil {
		return
	}

	releasePath := executablePath + "_" + version
	body, err := getFile(url + aoDownloadPath)
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

func getReleaseVersion(url string) (version string, err error) {
	releaseInfo, err := getFile(url + getCurrentVersionPath)
	if err != nil {
		return
	}
	releaseVersionStruct, err := versionutil.Json2Version(releaseInfo)

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
		err = errors.New(fmt.Sprintf("Error downloading update from %v: %v", url, err))
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
