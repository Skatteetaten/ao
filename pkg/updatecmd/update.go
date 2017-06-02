package updatecmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func UpdateSelf(args []string) (output string, err error) {
	executablePath, err := filepath.Abs(os.Args[0])

	if err != nil {
		return
	}
	fmt.Println(executablePath)

	const aoc5Url = "http://uil0map-hkldev-app01/aoc-v4/aoc"

	req, err := http.NewRequest(http.MethodGet, aoc5Url, nil)
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

	body, _ := ioutil.ReadAll(resp.Body)
	err = ioutil.WriteFile(executablePath + "_tmp", []byte(body), 0750)

	err = os.Rename(executablePath + "_tmp", executablePath)
	if err != nil {
		return
	}

	return
}
