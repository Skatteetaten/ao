package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

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
