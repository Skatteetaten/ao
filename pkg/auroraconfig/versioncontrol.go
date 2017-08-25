package auroraconfig

import (
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const GIT_URL_FORMAT = "https://%s@git.aurora.skead.no/scm/ac/%s.git"

func Clone(affiliation string, username string, outputPath string) error {

	url := fmt.Sprintf(GIT_URL_FORMAT, username, affiliation)
	fmt.Printf("Cloning AuroraConfig for affiliation %s\n", affiliation)
	fmt.Printf("%s\n\n", url)

	basicAuth := authenticateUser(username)

	_, err := git.PlainClone(outputPath, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     basicAuth,
	})

	if err != nil {
		return errors.Wrap(err, "Clone failed")
	}

	return nil
}

func Pull(username string) error {
	wd, _ := os.Getwd()

	repository, err := git.PlainOpen(wd)
	if err != nil {
		return err
	}

	basicAuth := authenticateUser(username)
	wt, _ := repository.Worktree()

	return wt.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       basicAuth,
	})

}

func Commit(username string, config *configuration.ConfigurationClass) error {

	wd, _ := os.Getwd()

	repository, err := git.PlainOpen(wd)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(GIT_URL_FORMAT, username, config.GetAffiliation())
	if err = validateRepo(url, repository); err != nil {
		return err
	}

	basicAuth := authenticateUser(username)
	// returns error if repository is already up to date
	fetchOrigin(repository, basicAuth)

	if err = checkForNewCommits(); err != nil {
		return err
	}

	if err = checkRepoForChanges(repository); err != nil {
		return err
	}

	if err = handleAuroraConfigCommit(repository, config); err != nil {
		return err
	}

	wt, _ := repository.Worktree()
	if err = wt.Checkout(&git.CheckoutOptions{Branch: "."}); err != nil {
		return err
	}

	return wt.Pull(&git.PullOptions{Auth: basicAuth})
}

func validateRepo(gitUrl string, repository *git.Repository) error {

	remote, _ := repository.Remote("origin")
	remoteUrl := remote.Config().URL

	if gitUrl != remoteUrl {
		message := fmt.Sprintf(`Wrong repository.
Expected remote to be %s
But was %s`, gitUrl, remoteUrl)
		return errors.New(message)
	}

	return nil
}

func authenticateUser(username string) *http.BasicAuth {
	fmt.Print("Enter password: ")
	password, _ := gopass.GetPasswdMasked()

	fmt.Println()

	return http.NewBasicAuth(username, string(password))
}

func handleAuroraConfigCommit(repository *git.Repository, config *configuration.ConfigurationClass) error {

	ac, err := GetAuroraConfig(config)

	if err != nil {
		return errors.Wrap(err, "Failed getting AuroraConfig")
	}

	if err = addFilesToAuroraConfig(&ac); err != nil {
		return errors.Wrap(err, "Failed adding files to AuroraConfig")
	}

	head, _ := repository.Head()
	removeFilesFromAuroraConfig(repository, &ac, head.Hash())

	if err = PutAuroraConfig(ac, config); err != nil {
		return errors.Wrap(err, "Failed committing AuroraConfig")
	}

	return nil
}

func checkRepoForChanges(repository *git.Repository) error {
	wt, _ := repository.Worktree()
	status, _ := wt.Status()
	if status.IsClean() {
		return errors.New("Nothing to commit")
	}

	return nil
}

func fetchOrigin(repository *git.Repository, auth *http.BasicAuth) error {

	return repository.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
	})
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

	cmd := exec.Command("git", "log", compare, "--oneline")
	out, err := cmd.Output()

	if err != nil {
		return err
	}

	output := string(out)

	if len(output) > 0 {
		return errors.New("new commits")
	}

	return nil
}

func addFilesToAuroraConfig(ac *serverapi_v2.AuroraConfig) error {
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

func removeFilesFromAuroraConfig(repository *git.Repository, ac *serverapi_v2.AuroraConfig, hash plumbing.Hash) error {

	wt, _ := repository.Worktree()
	status, _ := wt.Status()
	commit, _ := repository.CommitObject(hash)

	headFiles, _ := commit.Files()
	return headFiles.ForEach(func(file *object.File) error {
		code := status.File(file.Name).Worktree

		if code == git.Deleted {
			delete(ac.Files, file.Name)
		}

		return nil
	})
}
