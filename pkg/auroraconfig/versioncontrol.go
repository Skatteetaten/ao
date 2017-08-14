package auroraconfig

import (
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"strings"
	"os/exec"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"io/ioutil"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"path/filepath"
	"github.com/skatteetaten/ao/pkg/configuration"
)

func authenticateUser(username string) *http.BasicAuth {
	fmt.Printf("Enter password: ")
	password, _ := gopass.GetPasswdMasked()

	fmt.Println()

	return http.NewBasicAuth(username, string(password))
}

func Clone(affiliation string, username string, outputPath string, url string) (error) {

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

	currentDir, _ := os.Getwd()
	os.Chdir(outputPath)

	cmd := exec.Command("git", "branch", "--set-upstream-to=origin/master", "master")
	if err = cmd.Run(); err != nil {
		return err
	}

	os.Chdir(currentDir)

	return nil
}

func CheckRepo() error {

	if err := compareGitLog("origin/master..HEAD"); err != nil {
		return errors.Wrap(err, "run git reset HEAD")
	}
	if err := compareGitLog("HEAD..origin/master"); err != nil {
		return errors.Wrap(err, "run git pull")
	}

	return nil
}

func Commit(username string, persistentOptions *cmdoptions.CommonCommandOptions) error {

	wd, err := os.Getwd()

	repository, err := git.PlainOpen(wd)
	if err != nil {
		return err
	}

	basicAuth := authenticateUser(username)
	if err = fetchOrigin(repository, basicAuth); err != nil {
		fmt.Println(err)
	}

	if err = CheckRepo(); err != nil {
		return err
	}

	var config configuration.ConfigurationClass
	config.Init(persistentOptions)

	ac, err := GetAuroraConfig(&config)

	if err != nil {
		return err
	}

	if err = addFilesToAuroraConfig(&ac); err != nil {
		fmt.Println(err)
	}

	head, _ := repository.Head()
	removeFilesFromAuroraConfig(repository, &ac, head.Hash())

	if err = PutAuroraConfig(ac, &config); err != nil {
		return err
	}

	if err = exec.Command("git", "checkout", ".").Run(); err != nil {
		return err
	}

	wt, _ := repository.Worktree()

	wt.Pull(&git.PullOptions{
		Auth:     basicAuth,
		Progress: os.Stdout,
	})

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

func fetchOrigin(repository *git.Repository, auth *http.BasicAuth) error {

	return repository.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
	})
}

/**
 * git log HEAD..origin/master
 */
func IsRepositoryUpToDate(branch string, repository *git.Repository) error {

	refs, _ := repository.References()
	originHash := getBranchHash(branch, refs)
	head, _ := repository.Head()

	headLog, _ := repository.Log(&git.LogOptions{
		From: head.Hash(),
	})

	originLog, _ := repository.Log(&git.LogOptions{
		From: originHash,
	})

	var newCommits []string
	headCommit, _ := headLog.Next()

	originLog.ForEach(func(commit *object.Commit) error {

		if headCommit.Hash != commit.Hash {
			newCommits = append(newCommits, commit.Hash.String())
		}

		originLog.Close()

		return nil
	})

	if len(newCommits) > 0 {
		return errors.New("You need to pull the latest configuration. (git pull)")
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
		fmt.Printf(output)
		return errors.New("new commits")
	}

	return nil
}

func resetTo(branch string, refs storer.ReferenceIter, headHash plumbing.Hash) error {

	branchHash := getBranchHash(branch, refs)

	if branchHash.String() == headHash.String() {
		return nil
	}

	gitReset := exec.Command("git", "reset", branchHash.String())

	fmt.Println("Reset commits to " + branch)

	err := gitReset.Run()
	if err != nil {
		return err
	}

	return nil
}

func getBranchHash(branch string, refs storer.ReferenceIter) plumbing.Hash {

	branchHash := plumbing.NewHash("")

	refs.ForEach(func(ref *plumbing.Reference) error {

		if strings.Contains(ref.Name().String(), branch) {
			branchHash = ref.Hash()
		}

		return nil
	})

	return branchHash
}
