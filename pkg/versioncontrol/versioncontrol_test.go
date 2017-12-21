package versioncontrol

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var testFiles = map[string]string{
	"about.json": `{
		"schemaVersion": "v1",
		"affiliation": "paas",
		"permissions": {
		  "admin": "test"
		}
	}`,
	"reference.json": `{
		"artifactId": "openshift-reference-springboot-server",
		"groupId": "no.skatteetaten.aurora.openshift",
		"version": "1",
		"route": true
	}`,
	"test/about.json": `{
		"cluster": "qa"
	}`,
	"test/reference.json": `{}`,
}

const REPO_PATH = "/tmp/ao/testRepo"
const GIT_URL_FORMAT = "https://%s@git.aurora.skead.no/scm/ac/%s.git"

func repoSetup(gitRemoteUrl string) {
	// Clear old test files
	os.RemoveAll(REPO_PATH)
	os.MkdirAll(REPO_PATH, 0755)
	os.Chdir(REPO_PATH)

	if err := exec.Command("git", "init").Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("git", "remote", "add", "origin", gitRemoteUrl).Run(); err != nil {
		panic(err)
	}
}

func TestFindGitPath(t *testing.T) {
	gitRemoteUrl := fmt.Sprintf(GIT_URL_FORMAT, "user", "aurora")
	repoSetup(gitRemoteUrl)

	test := REPO_PATH + "/random/test"

	os.MkdirAll(test, 0755)
	os.Chdir(test)

	wd, _ := os.Getwd()
	path, err := FindGitPath(wd)
	if err != nil || path != REPO_PATH {
		t.Error("Expected git repo to be found")
	}
}
