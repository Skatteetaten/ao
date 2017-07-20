package updatecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/versionutil"
	"io/ioutil"
	"net/http"
	"os"
)

const aoc5Url = "http://aoc-update-service-paas-aoc-update.utv.paas.skead.no"

func UpdateSelf(args []string, simulate bool, forceVersion string, forceUpdate bool) (output string, err error) {
	var releaseVersion string
	if err != nil {
		return
	}

	if forceVersion == "" {
		releaseVersion, err = getReleaseVersion()
		if err != nil {
			return
		}
	} else {
		releaseVersion = forceVersion
	}

	myVersion, err := getMyVersion()

	if myVersion != releaseVersion || forceUpdate {
		output += "New version detected: Current version: " + myVersion + ".  Available version: " + releaseVersion
		if !simulate {
			err = doUpdate(releaseVersion)
			if err != nil {
				return
			}
			output += "\nAOC updated sucessfully"
		}
	} else {
		output += "No update available"
	}

	return
}

func doUpdate(version string) (err error) {
	releaseFilename := "aoc_" + version
	releaseUrl := aoc5Url + "/" + releaseFilename

	executablePath, err := os.Executable()
	if err != nil {
		return
	}

	releasePath := executablePath + "_" + version
	fmt.Println("DEBUG: Executable path: " + executablePath)
	fmt.Println("DEBUG: Release path: " + releasePath)
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
	releaseinfoUrl := aoc5Url + "/releaseinfo.json"
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

		err = errors.New(fmt.Sprintf("Error downloading update from %v: %v", aoc5Url, err))
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
