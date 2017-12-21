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

func FindGitPath(path string) (string, error) {
	current := fmt.Sprintf("%s/.git", path)
	if _, err := os.Stat(current); err == nil {
		return path, nil
	}

	paths := strings.Split(path, "/")
	length := len(paths)
	if length == 1 {
		return "", errors.New("Not a git repository")
	}

	next := strings.Join(paths[:length-1], "/")
	return FindGitPath(next)
}

func CollectJSONFilesInRepo(gitRoot string) (map[string]json.RawMessage, error) {
	files := make(map[string]json.RawMessage)
	return files, filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {

		fileName := strings.TrimPrefix(path, gitRoot+"/")

		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".json") {
			return nil
		}

		file, err := ioutil.ReadFile(gitRoot + "/" + fileName)

		if err != nil {
			return errors.Wrap(err, "Could not read file "+fileName)
		}

		if !json.Valid(file) {
			err = errors.New("Illegal JSON in file " + fileName)
			return err
		}

		files[fileName] = file
		return nil
	})
}
