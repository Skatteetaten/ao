package auroraconfig

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/openshift"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

const GIT_URL_FORMAT = "https://%s@git.aurora.skead.no/scm/ac/%s.git"

// TODO: Needs testing
// TODO: Add debug

func GitCommand(args ...string) (string, error) {
	command := exec.Command("git", args...)

	cmdReader, err := command.StdoutPipe()
	cmdErrReader, err := command.StderrPipe()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(cmdReader)
	errScanner := bufio.NewScanner(cmdErrReader)

	err = command.Start()
	if err != nil {
		return "", errors.Wrap(err, "Failed to start git command")
	}

	errMessage := ""
	for errScanner.Scan() {
		errMessage = fmt.Sprintf("%s%s", errMessage, errScanner.Text())
	}

	message := ""
	for scanner.Scan() {
		message = fmt.Sprintf("%s%s\n", message, scanner.Text())
	}

	err = command.Wait()
	if err != nil {
		return "", errors.New(errMessage)
	}

	return message, nil
}

func Checkout(url string, outputPath string) (string, error) {
	return GitCommand("clone", url, outputPath)
}

func Pull() (string, error) {
	statuses, err := getStatuses()
	if err != nil {
		return "", err
	}

	if len(statuses) == 0 {
		return GitCommand("pull")
	}

	if _, err := GitCommand("stash"); err != nil {
		return "", err
	}
	if _, err := GitCommand("pull"); err != nil {
		return "", err
	}
	if _, err := GitCommand("stash", "pop"); err != nil {
		return "", err
	}

	return "", nil
}

func getStatuses() ([]string, error) {
	var statuses []string
	if status, err := GitCommand("status", "-s"); err != nil {
		return statuses, errors.Wrap(err, "Failed to get status from repo")
	} else {
		statuses = strings.Fields(status)
	}

	return statuses, nil
}

func Save(url string, config *configuration.ConfigurationClass) (string, error) {
	if err := ValidateRepo(url); err != nil {
		return "", err
	}

	statuses, err := getStatuses()
	if err != nil {
		return "", err
	}

	if !isCleanRepo() {
		fetchOrigin()
		if err := checkForNewCommits(); err != nil {
			return "", err
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

// TODO: Test
func Validate(config *configuration.ConfigurationClass) (string, []string, error) {
	auroraConfig := &serverapi.AuroraConfig{
		Files:    make(map[string]json.RawMessage),
		Versions: make(map[string]string),
	}

	if err := addFilesToAuroraConfig(auroraConfig); err != nil {
		return "", []string{}, err
	}

	return ValidateAuroraConfig(auroraConfig, config)
}

func isCleanRepo() bool {
	_, err := GitCommand("log", "-1")
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

	extractAffiliation := func(url string) string {
		split := strings.Split(url, "/")
		length := len(split)
		if length == 0 {
			return ""
		}
		return strings.TrimSuffix(split[length-1], ".git")
	}

	remotes := strings.Fields(output)
	var repoUrl string
	for i, v := range remotes {
		if v == "origin" && len(remotes) > i+1 {
			repoUrl = remotes[i+1]
			break
		}
	}

	expectedAffiliation := extractAffiliation(expectedUrl)
	repoAffiliation := extractAffiliation(repoUrl)

	if expectedAffiliation != repoAffiliation {
		message := fmt.Sprintf(`Wrong repository.
Expected affliation to be %s, but was %s.`, expectedAffiliation, repoAffiliation)
		return errors.New(message)
	}

	return nil
}

func handleAuroraConfigCommit(statuses []string, config *configuration.ConfigurationClass) error {
	// TODO: Remove this request
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

func addFilesToAuroraConfig(ac *serverapi.AuroraConfig) error {

	wd, _ := os.Getwd()
	gitRoot, found := FindGitPath(wd)
	if !found {
		return errors.New("Could not find git")
	}

	return filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {

		filename := strings.TrimPrefix(path, gitRoot+"/")

		if strings.Contains(filename, ".git") || strings.Contains(filename, ".secret") || info.IsDir() {
			return nil
		}

		file, err := ioutil.ReadFile(gitRoot + "/" + filename)

		if err != nil {
			return errors.Wrap(err, "Could not read file "+filename)
		}

		if !jsonutil.IsLegalJson(string(file)) {
			err = errors.New("Illegal JSON in file " + filename)
			return err
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
