package versioncontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestValidateRepo(t *testing.T) {
	gitRemoteUrl := fmt.Sprintf(GIT_URL_FORMAT, "user", "aurora")
	repoSetup(gitRemoteUrl)

	if err := ValidateRepo(gitRemoteUrl); err != nil {
		t.Error(err)
	}
}

func TestFindGitPath(t *testing.T) {
	gitRemoteUrl := fmt.Sprintf(GIT_URL_FORMAT, "user", "aurora")
	repoSetup(gitRemoteUrl)

	test := REPO_PATH + "/random/test"

	os.MkdirAll(test, 0755)
	os.Chdir(test)

	wd, _ := os.Getwd()
	path, found := FindGitPath(wd)
	if !found || path != REPO_PATH {
		t.Error("Expected git repo to be found")
	}
}

func TestGetStatuses(t *testing.T) {
	t.Run("Should get statuses from the root of the repository to prevent deletng the wrong files", func(t *testing.T) {
		gitRemoteURL := fmt.Sprintf("file:///tmp/boobergit/%s", "aurora")
		repoSetup(gitRemoteURL)

		// Add test files to repository
		for f, v := range testFiles {
			filePath := path.Join(REPO_PATH, f)
			if split := strings.Split(f, "/"); len(split) == 2 {
				dirPath := path.Join(REPO_PATH, split[0])
				os.MkdirAll(dirPath, 0755)
			}
			fmt.Println(filePath)
			ioutil.WriteFile(filePath, []byte(v), 0644)
		}

		// Commit test files
		if err := exec.Command("git", "add", "--all").Run(); err != nil {
			t.Fatal(err)
		}
		if err := exec.Command("git", "commit", "-m", "init").Run(); err != nil {
			t.Fatal(err)
		}

		// Remove test file
		fileToRemove := "test/reference.json"
		os.Remove(path.Join(REPO_PATH, fileToRemove))

		// Change dir to where test file has been removed from
		testDir := path.Join(REPO_PATH, "test")
		os.Chdir(testDir)

		// Get statuses from repository
		statuses, err := getStatuses()
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, "D", statuses[0])
		assert.Equal(t, fileToRemove, statuses[1])
	})
}
