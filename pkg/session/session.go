package session

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"io/ioutil"
	"strings"
)

// SessionFileLocation is the location of the file holding session data for the login session
var SessionFileLocation string

// AOSession is a structure of a logged in session for ao
type AOSession struct {
	RefName      string            `json:"refName"`
	APICluster   string            `json:"apiCluster"`
	AuroraConfig string            `json:"auroraConfig"`
	Localhost    bool              `json:"localhost"`
	Tokens       map[string]string `json:"tokens"`
}

// ClusterToken holds information of an Openshift cluster including login token
type ClusterToken struct {
	Token string `json:"token"`
}

func LoadOrCreateAOSessionFile(sessionFileLocation string, aoConfig *config.AOConfig) (*AOSession, error) {
	aoSession, err := LoadSessionFile(sessionFileLocation)
	if err != nil {
		logrus.Debugln("Could not load session file.  Not logged in.")
	}

	if aoSession == nil {
		logrus.Info("Creating session file")
		aoSession = &AOSession{
			RefName:      "master",
			APICluster:   aoConfig.SelectAPICluster(),
			AuroraConfig: "",
			Localhost:    false,
			Tokens:       map[string]string{},
		}
		WriteAOSession(*aoSession, sessionFileLocation)
	} else if len(strings.TrimSpace(aoSession.APICluster)) == 0 {
		aoSession.APICluster = aoConfig.SelectAPICluster()
		WriteAOSession(*aoSession, sessionFileLocation)
		logrus.Info("Auto-selected apicluster.")
	}
	return aoSession, nil
}

// LoadSessionFile loads the login session file from file system
func LoadSessionFile(sessionFileLocation string) (*AOSession, error) {
	raw, err := ioutil.ReadFile(sessionFileLocation)
	if err != nil {
		return nil, err
	}

	var aoSession *AOSession
	err = json.Unmarshal(raw, &aoSession)
	if err != nil {
		return nil, err
	} else {
		aoSession.APICluster = strings.TrimSpace(aoSession.APICluster)
	}

	return aoSession, nil
}

// WriteAOSession writes the login session file to file system
func WriteAOSession(ao AOSession, sessionFileLocation string) error {
	data, err := json.MarshalIndent(ao, "", "  ")
	if err != nil {
		return fmt.Errorf("While marshaling ao session: %w", err)
	}
	if err := ioutil.WriteFile(sessionFileLocation, data, 0644); err != nil {
		return fmt.Errorf("While writing ao session to file: %w", err)
	}

	return nil
}
