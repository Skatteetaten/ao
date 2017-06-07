package updatecmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/versionutil"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const aoc5Url = "http://uil0map-hkldev-app01/aoc-v5"

func UpdateSelf(args []string, simulate bool, forceVersion string) (output string, err error) {
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

	if myVersion != releaseVersion {
		fmt.Println("DEBUG: New version detected")
		if simulate {
			output = "Update to " + releaseVersion
		} else {
			doUpdate(releaseVersion)
		}
	}
	
	return
}

func doUpdate (version string) (err error) {
	releaseFilename := "aoc_" + version
	releaseUrl := aoc5Url + "/" + releaseFilename

	executablePath, err := filepath.Abs(os.Args[0])
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
	releaseinfoUrl := aoc5Url + "/releaseinfo.json"
	releaseinfo, err := getFile(releaseinfoUrl)
	if err != nil {
		return
	}
	fmt.Println(string(releaseinfo))
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
