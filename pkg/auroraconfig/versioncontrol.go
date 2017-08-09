package auroraconfig

import (
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"github.com/pkg/errors"
)

func Clone(affiliation string, username string, path string) (error) {
	url := fmt.Sprintf("https://%s@git.aurora.skead.no/scm/ac/%s.git", username, affiliation)

	fmt.Printf("Cloning AuroraConfig for affiliation %s\n", affiliation)
	fmt.Printf("%s\n\n", url)

	fmt.Printf("Enter password: ")
	password, _ := gopass.GetPasswdMasked()

	fmt.Println()

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     http.NewBasicAuth(username, string(password)),
	})

	if err != nil {
		return errors.Wrap(err, "Clone failed")
	}

	return nil
}
