package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
	aoDownloadPath       = "/assets/ao"
	aoCurrentVersionPath = "/assets/version.json"
)

type AOVersion struct {
	Version    string `json:"version"`
	Branch     string `json:"branch"`
	GitHash    string `json:"gitHash"`
	BuildStamp string `json:"buildStamp"`
}

func (v *AOVersion) IsNewVersion() bool {
	// No new version if current version is dirty
	if strings.Contains(Version, "-dirty") {
		return false
	}
	// TODO: Should do better check then this
	return v.Version != Version
}

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

func GetNewAOClient(url string) ([]byte, error) {
	return fetchFromUpdateServer(url, aoDownloadPath, "application/octet-stream")
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
