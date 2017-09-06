package auroraconfig

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

const GIT_URL_FORMAT = "https://%s@git.aurora.skead.no/scm/ac/%s.git"

// TODO: Add debug
func GitCommand(args ...string) (string, error) {
	command := exec.Command("git", args...)
	cmdReader, err := command.StdoutPipe()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(cmdReader)

	err = command.Start()
	if err != nil {
		return "", errors.Wrap(err, "Failed to start git command")
	}

	message := ""
	for scanner.Scan() {
		message = fmt.Sprintf("%s%s\n", message, scanner.Text())
	}

	err = command.Wait()
	if err != nil {
		return "", errors.Wrap(err, "Failed to wait for git command")
	}

	return message, nil
}

func Checkout(url string, outputPath string) (string, error) {
	return GitCommand("clone", url, outputPath)
}

func Pull() (string, error) {
	if output, err := GitCommand("pull"); err != nil {
		return "", errors.Wrap(err, "Failed to pull new config")
	} else {
		return output, nil
	}
}

func Save(url string, config *configuration.ConfigurationClass) (string, error) {
	if err := ValidateRepo(url); err != nil {
		return "", errors.Wrap(err, "Failed to validate repo")
	}

	var statuses []string
	if status, err := GitCommand("status", "-s"); err != nil {
		return "", errors.Wrap(err, "Failed to get status from repo")
	} else {
		statuses = strings.Fields(status)
	}

	if !isCleanRepo() {
		fetchOrigin()
		if err := checkForNewCommits(); err != nil {
			return "", errors.Wrap(err, "Failed to check for new commits")
		}
	}

	if err := checkRepoForChanges(statuses); err != nil {
		return "", err
	}

	if err := handleAuroraConfigCommit(statuses, config); err != nil {
		return "", errors.Wrap(err, "Failed to save AuroraConfig")
	}

	// Delete untracked files
	if _, err := GitCommand("clean", "-fd"); err != nil {
		return "", errors.Wrap(err, "Failed to delete untracked files")
	}

	// Reset branch before pull
	if _, err := GitCommand("reset", "--hard"); err != nil {
		return "", errors.Wrap(err, "Failed to clean repo")
	}

	return Pull()
}

func isCleanRepo() bool {
	_, err := GitCommand("log")
	if err != nil {
		return true
	}

	return false
}

func UpdateLocalRepository(affiliation string, config *openshift.OpenshiftConfig) error {
	path := config.CheckoutPaths[affiliation]
	if path == "" {
		return errors.New("No local repository for affiliation " + affiliation)
	}

	wd, _ := os.Getwd()
	if err := os.Chdir(path); err != nil {
		return err
	}

	if _, err := Pull(); err != nil {
		return err
	}

	return os.Chdir(wd)
}

func ValidateRepo(expectedUrl string) error {
	output, err := GitCommand("remote", "-v")
	if err != nil {
		return err
	}

	remotes := strings.Fields(output)

	var repoUrl string
	for i, v := range remotes {
		if v == "origin" && len(remotes) > i+1 {
			repoUrl = remotes[i+1]
			break
		}
	}

	if repoUrl != expectedUrl {
		message := fmt.Sprintf(`Wrong repository.
Expected remote to be %s, actual %s.`, expectedUrl, repoUrl)
		return errors.New(message)
	}

	return nil
}

func handleAuroraConfigCommit(statuses []string, config *configuration.ConfigurationClass) error {
	ac, err := GetAuroraConfig(config)

	if err != nil {
		return errors.Wrap(err, "Failed getting AuroraConfig")
	}

	if err = addFilesToAuroraConfig(&ac); err != nil {
		return errors.Wrap(err, "Failed adding files to AuroraConfig")
	}

	removeFilesFromAuroraConfig(statuses, &ac)

	if err = PutAuroraConfig(ac, config); err != nil {
		return errors.Wrap(err, "Failed committing AuroraConfig")
	}

	return nil
}

func checkRepoForChanges(statuses []string) error {
	if len(statuses) == 0 {
		return errors.New("Nothing to save")
	}

	return nil
}

func fetchOrigin() (string, error) {

	return GitCommand("fetch", "origin")
}

func checkForNewCommits() error {

	if err := compareGitLog("origin/master..HEAD"); err != nil {
		return errors.New(`You have committed local changes.
Please revert them with: git reset HEAD^`)
	}

	if err := compareGitLog("HEAD..origin/master"); err != nil {
		return errors.New(`Please update to latest configuration with: ao pull`)
	}

	return nil
}

func compareGitLog(compare string) error {
	output, err := GitCommand("log", compare, "--oneline")
	if err != nil {
		return err
	}

	if len(output) > 0 {
		return errors.New("new commits")
	}

	return nil
}

func addFilesToAuroraConfig(ac *serverapi.AuroraConfig) error {
	wd, _ := os.Getwd()

	return filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {

		filename := strings.TrimPrefix(path, wd+"/")

		if strings.Contains(filename, ".git") || strings.Contains(filename, ".secret") || info.IsDir() {
			return nil
		}

		file, err := ioutil.ReadFile(wd + "/" + filename)

		if err != nil {
			return errors.Wrap(err, "Could not read file "+filename)
		}

		ac.Files[filename] = file

		return nil
	})
}

func removeFilesFromAuroraConfig(statuses []string, ac *serverapi.AuroraConfig) error {
	for i, v := range statuses {
		if v == "D" && len(statuses) > i+1 {
			delete(ac.Files, statuses[i+1])
		}
	}
	return nil
}
