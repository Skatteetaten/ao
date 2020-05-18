package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// The version variables will be set during build time, see build/build.sh
var (
	BuildStamp string
	Branch     string
	GitHash    string
	Version    string

	DefaultAOVersion = AOVersion{
		Version:    Version,
		BuildStamp: BuildStamp,
		Branch:     Branch,
		GitHash:    GitHash,
	}
)

const (
	aoDownloadPath        = "/assets/ao"
	aoDownloadPathMacOs   = "/assets/macos/ao"
	aoDownloadPathWindows = "/assets/windows/ao.exe"
	aoCurrentVersionPath  = "/assets/version.json"
)

// AOVersion holds version info for ao
type AOVersion struct {
	Version    string `json:"version"`
	Branch     string `json:"branch"`
	GitHash    string `json:"gitHash"`
	BuildStamp string `json:"buildStamp"`
}

// IsNewVersion checks if version is new
func (v *AOVersion) IsNewVersion() bool {
	// No new version if current version is dirty
	if strings.Contains(Version, "-dirty") {
		return false
	}
	// TODO: Should do better check then this
	return v.Version != Version
}

// GetCurrentVersionFromServer gets current (latest release) version from server
func GetCurrentVersionFromServer(url string) (*AOVersion, error) {
	data, err := fetchFromUpdateServer(url, aoCurrentVersionPath, "application/json")
	if err != nil {
		return nil, err
	}

	var aoVersion AOVersion
	err = json.Unmarshal(data, &aoVersion)
	if err != nil {
		return nil, err
	}

	return &aoVersion, nil
}

// GetNewAOClient downloads a new ao client from update server
func GetNewAOClient(url string) ([]byte, error) {
	var downloadPath string
	downloadPath = aoDownloadPath
	if runtime.GOOS == "darwin" {
		downloadPath = aoDownloadPathMacOs
	}
	if runtime.GOOS == "windows" {
		downloadPath = aoDownloadPathWindows
	}
	return fetchFromUpdateServer(url, downloadPath, "application/octet-stream")
}

func fetchFromUpdateServer(url, endpoint, contentType string) ([]byte, error) {
	logrus.WithField("url", url).WithField("endpoint", endpoint).Info("Request")
	req, err := http.NewRequest(http.MethodGet, url+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	logrus.WithField("url", url).WithField("status", res.StatusCode).Info("Response")
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Cannot reach update server: " + res.Status)
	}

	defer res.Body.Close()
	file, err := ioutil.ReadAll(res.Body)

	return file, err
}
