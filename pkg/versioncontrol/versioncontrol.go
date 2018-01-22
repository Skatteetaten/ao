package versioncontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skatteetaten/ao/pkg/client"

	"encoding/json"

	"github.com/pkg/errors"
)

func Checkout(url string, outputPath string) error {
	cmd := exec.Command("git", "clone", url, outputPath)
	return cmd.Run()
}

func CreateGitValidateHook(gitPath, hookType, auroraConfig string) error {
	hookScript := "#!/bin/bash\nexec ao validate -a " + auroraConfig
	gitHookFile := fmt.Sprintf("%s/.git/hooks/%s", gitPath, hookType)
	err := ioutil.WriteFile(gitHookFile, []byte(hookScript), 0755)
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
	separator := string(filepath.Separator)
	current := filepath.Join(path, ".git")
	if _, err := os.Stat(current); err == nil {
		return path, nil
	}

	paths := strings.Split(path, separator)
	length := len(paths)
	if length == 1 {
		return "", errors.New("Not a git repository")
	}

	next := strings.Join(paths[:length-1], separator)
	return FindGitPath(next)
}

func CollectJSONFilesInRepo(affiliation, gitRoot string) (*client.AuroraConfig, error) {
	ac := &client.AuroraConfig{
		Name: affiliation,
	}
	return ac, filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {

		fileName := strings.TrimPrefix(path, gitRoot+string(filepath.Separator))
		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".json") {
			return nil
		}

		file, err := ioutil.ReadFile(filepath.Join(gitRoot, fileName))
		if err != nil {
			return errors.Wrap(err, "Could not read file "+fileName)
		}

		if !json.Valid(file) {
			err = errors.New("Illegal JSON in file " + fileName)
			return err
		}

		ac.Files = append(ac.Files, client.AuroraConfigFile{
			Name:     fileName,
			Contents: string(file),
		})
		return nil
	})
}
