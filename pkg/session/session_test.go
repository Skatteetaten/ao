package session

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const sessionTmpFile = "/tmp/ao-session_test.json"

func TestLoadSessionFile(t *testing.T) {
	defer os.Remove(sessionTmpFile)
	aoSession, _ := LoadSessionFile(sessionTmpFile)
	assert.Empty(t, aoSession)

	aoSession = &AOSession{
		RefName:      "master",
		APICluster:   "testApiCluster",
		AuroraConfig: "",
		Localhost:    false,
		Tokens:       map[string]string{},
	}

	WriteAOSession(*aoSession, sessionTmpFile)

	aoSession, _ = LoadSessionFile(sessionTmpFile)
	assert.NotEmpty(t, aoSession)
}
