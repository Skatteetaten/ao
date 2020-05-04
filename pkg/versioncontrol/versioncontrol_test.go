package versioncontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

const RepoPath = "/tmp/ao/testRepo"
const GitURLFormat = "https://git.aurora.skead.no/scm/ac/%s.git"

func repoSetup(gitRemoteURL string) {
	// Clear old test files
	os.RemoveAll(RepoPath)
	os.MkdirAll(RepoPath, 0755)
	os.Chdir(RepoPath)

	if err := exec.Command("git", "init").Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("git", "remote", "add", "origin", gitRemoteURL).Run(); err != nil {
		panic(err)
	}
}

func TestFindGitPath(t *testing.T) {
	gitRemoteURL := fmt.Sprintf(GitURLFormat, "aurora")
	repoSetup(gitRemoteURL)

	test := RepoPath + "/random/test"

	os.MkdirAll(test, 0755)
	os.Chdir(test)

	wd, _ := os.Getwd()
	path, err := FindGitPath(wd)
	if err != nil || path != RepoPath {
		t.Error("Expected git repo to be found")
	}
}

func TestGetGitUrl(t *testing.T) {
	type args struct {
		affiliation   string
		user          string
		gitURLPattern string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should get git url for BitBucket",
			args: args{
				affiliation:   "pros",
				user:          "hans",
				gitURLPattern: GitURLFormat,
			},
			want: "https://hans@git.aurora.skead.no/scm/ac/pros.git",
		},
		{
			name: "Should get git url for local",
			args: args{
				affiliation:   "pros",
				user:          "hans",
				gitURLPattern: "file:///tmp/local-git/%s",
			},
			want: "file:///tmp/local-git/pros",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGitURL(tt.args.affiliation, tt.args.user, tt.args.gitURLPattern); got != tt.want {
				t.Errorf("GetGitURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectJSONFilesInRepo(t *testing.T) {
	repoSetup("")
	// Write test files to test repo
	for name, text := range testFiles {
		split := strings.Split(name, "/")
		if len(split) == 2 {
			os.Mkdir(fmt.Sprintf("%s/%s", RepoPath, split[0]), 0755)
		}
		err := ioutil.WriteFile(fmt.Sprintf("%s/%s", RepoPath, name), []byte(text), 0644)
		if err != nil {
			t.Error(err)
		}
	}

	tests := []struct {
		name    string
		gitRoot string
		want    int
		wantErr bool
	}{
		{
			name:    "Should collect all JSON files from folder successfully",
			gitRoot: RepoPath,
			want:    4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectAuroraConfigFilesInRepo("aurora", tt.gitRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectJSONFilesInRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Files) != tt.want {
				t.Errorf("CollectJSONFilesInRepo() = %v, want %v", len(got.Files), tt.want)
			}
		})
	}
}

func TestCreateGitValidateHook(t *testing.T) {
	repoSetup("")
	type args struct {
		gitPath      string
		hookType     string
		auroraConfig string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Should add pre-push hook to .git/hooks",
			args: args{
				gitPath:      RepoPath,
				auroraConfig: "pros",
				hookType:     "pre-push",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateGitValidateHook(tt.args.gitPath, tt.args.hookType, tt.args.auroraConfig); (err != nil) != tt.wantErr {
				t.Errorf("CreateGitValidateHook() error = %v, wantErr %v", err, tt.wantErr)
			}
			data, err := ioutil.ReadFile(fmt.Sprintf("%s/.git/hooks/pre-push", RepoPath))
			assert.NotEmpty(t, data)
			assert.NoError(t, err)
		})
	}
}
