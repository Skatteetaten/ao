package versioncontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/pkg/errors"
)

func Checkout(url string, outputPath string) error {
	cmd := exec.Command("git", "clone", url, outputPath)
	return cmd.Run()
}

func CreatePreCommitHook(gitPath, auroraConfig string) error {
	preCommitHookScript := "#!/bin/bash\nexec ao validate -a " + auroraConfig
	gitHookFile := fmt.Sprintf("%s/.git/hooks/pre-commit", gitPath)
	err := ioutil.WriteFile(gitHookFile, []byte(preCommitHookScript), 0755)
	if err != nil {
		return err
	}

	return nil
}

func GetGitUrl(affiliation, user, gitUrlPattern string) string {
	if !strings.Contains(gitUrlPattern, "https://") {
		return fmt.Sprintf(gitUrlPattern, affiliation)
	}

	host := strings.TrimPrefix(gitUrlPattern, "https://")
	newPattern := fmt.Sprintf("https://%s@%s", user, host)
	return fmt.Sprintf(newPattern, affiliation)
}

func FindGitPath(path string) (string, bool) {
	current := fmt.Sprintf("%s/.git", path)
	if _, err := os.Stat(current); err == nil {
		return path, true
	}

	paths := strings.Split(path, "/")
	length := len(paths)
	if length == 1 {
		return "", false
	}

	next := strings.Join(paths[:length-1], "/")
	return FindGitPath(next)
}

func CollectFilesInRepo() (map[string]json.RawMessage, error) {
	wd, _ := os.Getwd()
	gitRoot, found := FindGitPath(wd)
	if !found {
		return nil, errors.New("Could not find git")
	}

	files := make(map[string]json.RawMessage)
	return files, filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {

		filename := strings.TrimPrefix(path, gitRoot+"/")

		if strings.HasPrefix(filename, ".") || !strings.HasSuffix(filename, ".json") {
			return nil
		}

		file, err := ioutil.ReadFile(gitRoot + "/" + filename)

		if err != nil {
			return errors.Wrap(err, "Could not read file "+filename)
		}

		if !json.Valid(file) {
			err = errors.New("Illegal JSON in file " + filename)
			return err
		}

		files[filename] = file
		return nil
	})
}
