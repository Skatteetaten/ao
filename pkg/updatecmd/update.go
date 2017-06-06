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

func UpdateSelf(args []string) (output string, err error) {
	executableDirectory := filepath.Dir(os.Args[0]) //
	executablePath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return
	}

	releaseinfoUrl := aoc5Url + "/releaseinfo.json"

	fmt.Println(executableDirectory)
	fmt.Println(executablePath)
	fmt.Println(releaseinfoUrl)

	releaseinfo, err := getFile(releaseinfoUrl)
	if err != nil {
		return
	}
	fmt.Println(string(releaseinfo))
	versionStruct, err := versionutil.Json2Version(releaseinfo)

	fmt.Println(versionStruct.BuildStamp)

	//body, err := getFile(aoc5Url)
	//err = ioutil.WriteFile(executablePath, []byte(body), 0750)

	return
}

func getReleaseVersion() (versionStruct versionutil.VersionStruct, err error) {
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
