package versioncontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ao/pkg/auroraconfig"

	"github.com/pkg/errors"
)

// Checkout performs a git clone of a repo (an aurora config repo)
func Checkout(url string, outputPath string) error {
	cmd := exec.Command("git", "clone", url, outputPath)
	return cmd.Run()
}

// CreateGitValidateHook creates a git validate
func CreateGitValidateHook(gitPath, hookType, auroraConfig string) error {
	hookScript := "#!/bin/bash\nexec ao validate -a " + auroraConfig
	gitHookFile := fmt.Sprintf("%s/.git/hooks/%s", gitPath, hookType)
	err := ioutil.WriteFile(gitHookFile, []byte(hookScript), 0755)
	if err != nil {
		return err
	}

	return nil
}

// GetGitURL gets git URL for an affiliation
func GetGitURL(affiliation, user, gitURLPattern string) string {
	if !strings.Contains(gitURLPattern, "https://") {
		return fmt.Sprintf(gitURLPattern, affiliation)
	}

	host := strings.TrimPrefix(gitURLPattern, "https://")
	newPattern := fmt.Sprintf("https://%s@%s", user, host)
	return fmt.Sprintf(newPattern, affiliation)
}

// FindGitPath searches for a valid git path in a given path
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

// CollectAuroraConfigFilesInRepo finds complete AuroraConfig for an affiliation
func CollectAuroraConfigFilesInRepo(affiliation, gitRoot string) (*auroraconfig.AuroraConfig, error) {
	ac := &auroraconfig.AuroraConfig{
		Name: affiliation,
	}

	return ac, filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {

		fileName := strings.TrimPrefix(path, gitRoot+string(filepath.Separator))

		if !HasOneOfExtension(fileName, []string{".json", ".yaml"}) {
			return nil
		}

		file, err := ioutil.ReadFile(filepath.Join(gitRoot, fileName))
		if err != nil {
			return errors.Wrap(err, "Could not read file "+fileName)
		}

		ac.Files = append(ac.Files, auroraconfig.File{
			Name:     fileName,
			Contents: string(file),
		})
		return nil
	})
}

// HasOneOfExtension checks text for an array of suffixes
func HasOneOfExtension(text string, items []string) bool {
	for _, item := range items {
		if ok := strings.HasSuffix(text, item); ok {
			return true
		}
	}
	return false
}
